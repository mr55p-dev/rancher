package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mr55p-dev/gonk"
)

type Ticket struct {
	Prefix string `config:"prefix,optional"`
	ID     string
}

func (t *Ticket) String() string {
	builder := new(strings.Builder)
	builder.WriteString(t.Prefix)
	builder.WriteString(t.ID)
	return builder.String()
}

type Request struct {
	Separator            string `config:"separator,optional"`
	Type                 string `config:"type,optional"`
	Description          string
	DescriptionSeparator string `config:"description-separator,optional"`
}

type BranchOption struct {
	Key   string `config:"key"`
	Value string `config:"value"`
}

type Config struct {
	Ticket        Ticket         `config:"ticket,optional"`
	Request       Request        `config:"request,optional"`
	BranchOptions []BranchOption `config:"branch-options,optional"`
}

func (c *Config) String() string {
	builder := new(strings.Builder)
	if c.Request.Type != "" {
		builder.WriteString(c.Request.Type)
		builder.WriteString(c.Request.Separator)
	}
	builder.WriteString(c.Ticket.String())
	if c.Request.Description != "" {
		builder.WriteString(c.Request.Separator)
		builder.WriteString(c.Request.Description)
	}
	return builder.String()
}

func NewConfig() *Config {
	return &Config{
		Ticket: Ticket{
			Prefix: "GO",
		},
		Request: Request{
			Separator:            "/",
			Type:                 "feat",
			DescriptionSeparator: "-",
		},
		BranchOptions: []BranchOption{
			{"Feature", "feat"},
			{"Fix", "fix"},
			{"Documentation", "docs"},
			{"Refactor", "refactor"},
			{"Performance", "perf"},
			{"CI", "ci"},
			{"None", ""},
		},
	}
}

func runGit(args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return
}

func getBranchOptions(options []BranchOption) []huh.Option[string] {
	out := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		out = append(out, huh.NewOption(option.Key, option.Value))
	}
	return out
}

func getTicketOptions() []huh.Option[string] {
	return []huh.Option[string]{}
}

func main() {
	config := NewConfig()
	yamlLoader, _ := gonk.NewYamlLoader("rancher.yml")
	err := gonk.LoadConfig(config, yamlLoader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	branchOpts := getBranchOptions(config.BranchOptions)
	branchType := config.Request.Type
	branchDesc := config.Request.Description

	ticketInput := huh.NewInput().
		Key("ticketNumber").
		Title("Ticket No").
		Prompt("? ").
		Validate(func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return err
			}
			return nil
		})

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("branchType").
				Title("Branch Type").
				Options(branchOpts...).
				Value(&branchType),
			ticketInput,
			huh.NewInput().
				Key("branchDesc").
				Title("Description").
				Prompt("? ").
				Value(&branchDesc),
		),
	)
	err = form.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	config.Ticket.ID = form.GetString("ticketNumber")
	config.Request.Type = form.GetString("branchType")
	config.Request.Description = form.GetString("branchDesc")

	branchOut := config.String()
	var create bool = true
	huh.NewConfirm().
		Affirmative("Create").
		Negative("Cancel").
		Title("Create branch?").
		Description(branchOut).
		Value(&create).
		Run()

	if create {
		runGit("branch", branchOut)
		runGit("switch", branchOut)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mr55p-dev/gonk"
)

var useJira = flag.Bool("jira", false, "Use Jira for ticket numbers")

type Ticket struct {
	Prefix string `config:"prefix,optional"`
	ID     string
}

func (t *Ticket) String() string {
	if t.ID == "" {
		return ""
	}
	builder := new(strings.Builder)
	if !strings.HasPrefix(t.ID, t.Prefix) {
		builder.WriteString(t.Prefix)
	}
	builder.WriteString(t.ID)
	return builder.String()
}

type Request struct {
	Separator            string `config:"separator,optional"`
	Type                 string `config:"type,optional"`
	Description          string
	DescriptionSeparator string `config:"description-separator,optional"`
}

type SelectOption struct {
	Key   string `config:"key"`
	Value string `config:"value"`
}

type Config struct {
	Ticket        Ticket         `config:"ticket,optional"`
	Request       Request        `config:"request,optional"`
	BranchOptions []SelectOption `config:"branch-options,optional"`
	Jira          Jira           `config:"jira,optional"`
}

var DefaultBranchOptions = []SelectOption{
	{"Feature", "feat"},
	{"Fix", "fix"},
	{"Documentation", "docs"},
	{"Refactor", "refactor"},
	{"Performance", "perf"},
	{"CI", "ci"},
	{"None", ""},
}

func ToHuh(options []SelectOption) []huh.Option[string] {
	out := make([]huh.Option[string], len(options))
	for i, elem := range options {
		out[i] = huh.NewOption(elem.Key, elem.Value)
	}
	return out
}

func (c *Config) String() string {
	builder := new(strings.Builder)
	if c.Request.Type != "" {
		builder.WriteString(c.Request.Type)
		builder.WriteString(c.Request.Separator)
	}
	if ticket := c.Ticket.String(); ticket != "" {
		builder.WriteString(ticket)
		builder.WriteString(c.Request.Separator)
	}
	if c.Request.Description != "" {
		replacer := strings.NewReplacer((" "), c.Request.DescriptionSeparator)
		builder.WriteString(replacer.Replace(c.Request.Description))
	}
	return builder.String()
}

func NewConfig() *Config {
	return &Config{
		Request: Request{
			Separator:            "/",
			Type:                 "feat",
			DescriptionSeparator: "-",
		},
	}
}

func (c *Config) ApplyBranchDefaults() {
	c.BranchOptions = append(c.BranchOptions, DefaultBranchOptions...)
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

func getBranchOptions(options []SelectOption) []huh.Option[string] {
	out := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		out = append(out, huh.NewOption(option.Key, option.Value))
	}
	return out
}

func getTicketOptions() []huh.Option[string] {
	return []huh.Option[string]{}
}

func loadConfig() (*Config, error) {
	config := NewConfig()
	baseConfigDir, _ := os.UserHomeDir()
	configPath := filepath.Join(baseConfigDir, ".config", "rancher", "rancher.yml")
	yamlLoader, _ := gonk.NewYamlLoader(configPath)
	err := gonk.LoadConfig(config, yamlLoader)
	if err != nil {
		log.Printf("hit an error: %+v, error: %v", *config, err)
		return nil, err
	}
	config.ApplyBranchDefaults()
	return config, nil
}

func main() {
	flag.Parse()

	config, err := loadConfig()
	if err != nil {
		log.Panicf("Error loading configuration: %v", err)
	}
	branchOpts := getBranchOptions(config.BranchOptions)
	branchType := config.Request.Type
	branchDesc := config.Request.Description

	var ticketInput huh.Field
	if *useJira == true {
		tickets, err := config.Jira.QueryTickets()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		ticketOptions := ToHuh(tickets)
		ticketInput = huh.NewSelect[string]().
			Key("ticketNumber").
			Title("Ticket").
			Options(ticketOptions...).
			Value(&branchType)
	} else {
		ticketInput = huh.NewInput().
			Key("ticketNumber").
			Title("Ticket No").
			Prompt("? ")
	}

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

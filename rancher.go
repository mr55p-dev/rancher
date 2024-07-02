package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/mr55p-dev/gonk"
)

var useJira = flag.Bool("jira", false, "Use Jira for ticket numbers")

func Git(args ...string) {
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

func getConfigDir() string {
	baseConfigDir, _ := os.UserHomeDir()
	return filepath.Join(baseConfigDir, ".config", "rancher", "rancher.yml")
}

func getConfig() (*Config, error) {
	config := NewConfig()
	configPath := getConfigDir()
	yamlLoader, _ := gonk.NewYamlLoader(configPath)
	err := gonk.LoadConfig(config, yamlLoader)
	if err != nil {
		log.Printf("hit an error: %+v, error: %v", *config, err)
		return nil, err
	}
	config.ApplyBranchDefaults()
	return config, nil
}

func getTicketInput(config *Config) huh.Field {
	var ticketInput huh.Field = huh.NewInput().
		Key("ticketNumber").
		Title("Ticket No").
		Value(&config.Ticket.ID).
		Prompt("? ")
	if *useJira == false {
		return ticketInput
	}

	tickets, err := config.Jira.QueryTickets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ticketInput
	}

	return huh.NewSelect[string]().
		Key("ticketNumber").
		Title("Ticket").
		Value(&config.Ticket.ID).
		Options(ToHuh(tickets)...)
}

func main() {
	flag.Parse()

	config, err := getConfig()
	if err != nil {
		log.Panicf("Error loading configuration: %v", err)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("branchType").
				Title("Branch Type").
				Options(ToHuh(config.BranchOptions)...).
				Value(&config.Branch.Type),
			getTicketInput(config),
			huh.NewInput().
				Key("branchDesc").
				Title("Description").
				Prompt("? ").
				Value(&config.Branch.Description),
		),
	)
	err = form.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var create bool = true
	huh.NewConfirm().
		Affirmative("Create").
		Negative("Cancel").
		Title("Create branch?").
		Description(config.String()).
		Value(&create).
		Run()

	if create {
		Git("branch", config.String())
		Git("switch", config.String())
	}
}

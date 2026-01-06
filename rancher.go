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
var debug = flag.Bool("debug", false, "Enable debug logging")
var doInit = flag.Bool("init", false, "Creates default config file at $HOME/.config/rancher/rancher.yml")

func Git(args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func getConfigDir() (string, error) {
	baseConfigDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Failed to get user home dir: %w", err)
	}
	return filepath.Join(baseConfigDir, ".config", "rancher"), nil
}

func getConfig() (*Config, error) {
	config := NewConfig()
	configDir, err := getConfigDir()
	if err != nil {
		log.Fatalln("Could not read user home directory", err.Error())
	}
	configPath := filepath.Join(configDir, "rancher.yml")
	baseYAMLLoader, _ := gonk.NewYamlLoader(configPath)
	localYAMLLoader, _ := gonk.NewYamlLoader(".rancher.yml")

	if err = gonk.LoadConfig(config, baseYAMLLoader, localYAMLLoader); err != nil {
		return nil, err
	}
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

	tickets, err := config.Jira.QueryTickets(*debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying Jira: %v\n", err)
		return ticketInput
	}

	return huh.NewSelect[string]().
		Key("ticketNumber").
		Title("Ticket").
		Value(&config.Ticket.ID).
		Options(tickets...)
}

func main() {
	flag.Parse()

	if *doInit {
		log.Println("Creating default config file")
		dir, err := getConfigDir()
		if err != nil {
			log.Fatalln("Could not get config dir name", err.Error())
		}
		if err := os.MkdirAll(dir, os.ModeDir); err != nil {
			log.Fatalln("Could not create config directory", err.Error())
		}
		tgt := filepath.Join(dir, "rancher.yml")
		if _, err := os.Stat(tgt); err == nil {
			log.Fatalln("Config file already exists. Please delete it to continue.")
		}
		fd, err := os.Create(tgt)
		if err != nil {
			log.Fatalln("Failed to create config file", err.Error())
		}
		if _, err := fd.WriteString(defaultConfig); err != nil {
			log.Fatalln("Failed to write to config file", err.Error())
		}
		log.Println("Succesfully created default config file")
		return
	}

	config, err := getConfig()
	if err != nil {
		log.Panicf("Error loading configuration: %v", err)
	}

	huhBranches := make([]huh.Option[string], 0, len(config.BranchTypeOptions))
	for _, entry := range config.BranchTypeOptions {
		for key, val := range entry {
			huhBranches = append(huhBranches, huh.Option[string]{
				Key:   key,
				Value: val,
			})
		}
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("branchType").
				Title("Branch Type").
				Options(huhBranches...).
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

	confirm := true
	_ = huh.NewConfirm().
		Affirmative("Create").
		Negative("Cancel").
		Title("Create branch?").
		Description(config.String()).
		Value(&confirm).
		Run()

	if confirm == true {
		Git("branch", config.String())
		Git("switch", config.String())
	}
}

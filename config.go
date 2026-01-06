package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed config.sample.yml
var defaultConfig string

type Ticket struct {
	ID string
}

type Branch struct {
	Type         string `config:"type,optional"`
	Description  string `config:"-"`
	FormatString string `config:"format-string"`
}

type FormatOptions struct {
	Type      string
	TypeShort string

	TicketId      string
	TicketNumber  string
	TicketProject string

	Description string
}

type Config struct {
	// Ticket info
	Ticket Ticket `config:"-"`
	// what choices are for the branch type field
	BranchTypeOptions []map[string]string `config:"types,optional"`
	// Branch name generation settings
	Branch Branch `config:"branch,optional"`
	// Jira API config
	Jira Jira `config:"jira,optional"`
}

func sanitize(s, separator string) string {
	replacer := strings.NewReplacer((" "), separator)
	return replacer.Replace(strings.TrimSpace(s))
}

func (c *Config) LoadOptions() FormatOptions {
	segments := strings.Split(c.Ticket.ID, "-")
	if len(segments) != 2 {
		segments = []string{"", ""}
	}

	return FormatOptions{
		Type:          c.Branch.Type,
		TypeShort:     c.Branch.Type,
		TicketId:      c.Ticket.ID,
		TicketProject: segments[0],
		TicketNumber:  segments[1],
		Description:   sanitize(c.Branch.Description, "-"),
	}
}

func (c *Config) String() string {
	tmpl, err := template.New("branch").Parse(c.Branch.FormatString)
	if err != nil {
		panic(err)
	}

	options := c.LoadOptions()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, options); err != nil {
		panic(err)
	}

	return buf.String()
}

func NewConfig() *Config {
	return &Config{
		Branch: Branch{
			Type:         "feat",
			FormatString: "{{.Type}}/{{.TicketId}}-{{.Description}}",
		},
		Jira: Jira{
			Query: "assignee = currentUser()",
		},
	}
}

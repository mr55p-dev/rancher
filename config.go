package main

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
)

//go:embed config.sample.yml
var defaultConfig string

type Ticket struct {
	ID string
}

type Branch struct {
	// The separating string
	Separator            string `config:"separator,optional"`
	Type                 string `config:"type,optional"`
	Description          string `config:"-"`
	DescriptionSeparator string `config:"description-separator,optional"`
	FormatString         string `config:"format-string"`
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
	BranchTypeOptions map[string]string `config:"types,optional"`
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
		panic("Invalid ticket format")
	}

	return FormatOptions{
		Type:          c.Branch.Type,
		TypeShort:     c.Branch.Type,
		TicketId:      c.Ticket.ID,
		TicketNumber:  segments[0],
		TicketProject: segments[1],
		Description:   c.Branch.Description,
	}
}

func (c *Config) String() string {
	segments := make([]string, 0, 4)
	branchType := sanitize(c.Branch.Type, c.Branch.DescriptionSeparator)
	if branchType != "" {
		segments = append(segments, branchType)
	}

	ticketId := sanitize(c.Ticket.ID, "-")
	if ticketId != "" {
		segments = append(segments, ticketId)
	}

	branchDesc := sanitize(c.Branch.Description, c.Branch.DescriptionSeparator)
	if branchDesc != "" {
		segments = append(segments, branchDesc)
	}

	if c.Jira.SuffixTicketID {
		expr := regexp.MustCompile("[0-9]+")
		idInt := expr.FindString(c.Ticket.ID)
		segments = append(segments, fmt.Sprintf("#%s", idInt))
	}

	return strings.Join(segments, c.Branch.Separator)
}

func NewConfig() *Config {
	return &Config{
		Branch: Branch{
			Separator:            "/",
			Type:                 "feat",
			DescriptionSeparator: "-",
		},
		Jira: Jira{
			Query: "assignee = currentUser()",
		},
	}
}

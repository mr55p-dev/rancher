package main

import "strings"

type Ticket struct {
	ID string
}

type Branch struct {
	// The separating string
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
	// Ticket info
	Ticket Ticket
	// what choices are for the branch type field
	BranchTypeOptions []SelectOption `config:"types,optional"`
	// Branch name generation settings
	Branch Branch `config:"branch,optional"`
	// Jira API config
	Jira Jira `config:"jira,optional"`
}

var DefaultBranchTypeOpts = []SelectOption{
	{"Feature", "feat"},
	{"Fix", "fix"},
	{"Documentation", "docs"},
	{"Refactor", "refactor"},
	{"Performance", "perf"},
	{"CI", "ci"},
	{"None", ""},
}

func sanitize(s, separator string) string {
	replacer := strings.NewReplacer((" "), separator)
	return replacer.Replace(strings.TrimSpace(s))
}

func (c *Config) String() string {
	segments := make([]string, 0)
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

func (c *Config) ApplyBranchDefaults() {
	c.BranchTypeOptions = append(c.BranchTypeOptions, DefaultBranchTypeOpts...)
}

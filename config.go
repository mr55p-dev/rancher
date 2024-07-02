package main

import "strings"

type Ticket struct {
	ID string
}

type Branch struct {
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
	Ticket        Ticket
	BranchOptions []SelectOption
	Branch        Branch `config:"request,optional"`
	Jira          Jira   `config:"jira,optional"`
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

func (c *Config) String() string {
	segments := make([]string, 0)
	if c.Branch.Type != "" {
		segments = append(segments, c.Branch.Type)
	}
	if c.Ticket.ID != "" {
		segments = append(segments, c.Ticket.ID)
	}
	if c.Branch.Description != "" {
		replacer := strings.NewReplacer((" "), c.Branch.DescriptionSeparator)
		segments = append(segments, replacer.Replace(c.Branch.Description))
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
	}
}

func (c *Config) ApplyBranchDefaults() {
	c.BranchOptions = append(c.BranchOptions, DefaultBranchOptions...)
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

var (
	ticketPrefix string = "GO"
	ticketSep    string = "-"
	ticketNo     string = fmt.Sprintf("%s%s", ticketPrefix, ticketSep)

	branchSep  string = "/"
	branchType string = "feat"

	desc    string
	descSep string = "-"
)

type BranchParams struct {
	BranchRaw string
	TicketRaw string
	TypeRaw   string
	DescRaw   string
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

func NewBranchParams(branch string) *BranchParams {
	out := &BranchParams{
		BranchRaw: branch,
	}

	portions := strings.Split(branch, "/")
	switch len(portions) {
	case 3:
		out.DescRaw = portions[2]
		fallthrough
	case 2:
		out.TicketRaw = portions[1]
		fallthrough
	case 1:
		out.TypeRaw = portions[0]
	}
	return out

}

func getBranchOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Feature", "feat"),
		huh.NewOption("Fix", "fix"),
		huh.NewOption("Documentation", "docs"),
		huh.NewOption("Refactor", "refactor"),
		huh.NewOption("Performance", "perf"),
		huh.NewOption("CI", "ci"),
		huh.NewOption("None", ""),
	}
}

func getTicketOptions() []huh.Option[string] {
	return []huh.Option[string]{}
}

func main() {
	branchOpts := getBranchOptions()
	branchTicket := getTicketOptions()
	var ticketInput huh.Field
	if len(branchTicket) > 0 {
		ticketInput = huh.NewSelect[string]().
			Title("Ticket").
			Options(branchTicket...).
			Value(&ticketNo)
	} else {
		ticketInput = huh.NewInput().
			Title("Ticket No").
			Prompt("? ").
			Value(&ticketNo)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Branch Type").
				Options(branchOpts...).
				Value(&branchType),
			ticketInput,
			huh.NewInput().
				Title("Description").
				Prompt("? ").
				Value(&desc),
		),
	)
	err := form.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	branchOut := new(strings.Builder)
	desc = strings.TrimSpace(desc)
	desc = strings.ReplaceAll(desc, " ", descSep)

	components := []string{branchType, ticketNo, desc}
	for idx, component := range components {
		if idx > 0 && component != "" {
			branchOut.WriteString(branchSep)
		}
		if component != "" {
			branchOut.WriteString(component)
		}
	}
	var create bool = true
	huh.NewConfirm().
		Affirmative("Create").
		Negative("Cancel").
		Title("Create branch?").
		Description(branchOut.String()).
		Value(&create).
		Run()

	if create {
		runGit("branch", branchOut.String())
		runGit("switch", branchOut.String())
	}
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
)

var (
	ticketPrefix string = "GO"
	ticketSep    string = "-"
	ticketNo     string

	branchSep  string = "/"
	branchType string

	desc    string
	descSep string = " - "
)

type BranchParams struct {
	BranchRaw string
	TicketRaw string
	TypeRaw   string
	DescRaw   string
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

func getBranchOptions(branchPrefix string) []huh.Option[string] {
	opts := []huh.Option[string]{
		huh.NewOption("Feature", "feat"),
		huh.NewOption("Bug", "bug"),
		huh.NewOption("Documentation", "docs"),
		huh.NewOption("Refactor", "refactor"),
		huh.NewOption("Performance", "perf"),
		huh.NewOption("CI", "ci"),
		huh.NewOption("None", ""),
	}

	slices.SortStableFunc(opts, func(i, j huh.Option[string]) int {
		if i.Value == branchPrefix {
			return -1
		}
		return 0
	})

	return opts
}

func main() {
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchReader := new(strings.Builder)
	branchCmd.Stdout = branchReader
	err := branchCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	branchInfo := NewBranchParams(branchReader.String())
	branchOpts := getBranchOptions(branchInfo.TypeRaw)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Branch Type").
				Options(branchOpts...),
		),
	)
	err = form.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}

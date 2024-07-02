package main

import "github.com/charmbracelet/huh"

func ToHuh(options []SelectOption) []huh.Option[string] {
	out := make([]huh.Option[string], len(options))
	for i, elem := range options {
		out[i] = huh.NewOption(elem.Key, elem.Value)
	}
	return out
}

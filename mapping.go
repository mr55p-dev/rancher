package main

import "github.com/charmbracelet/huh"

func ToHuh(options map[string]string) []huh.Option[string] {
	out := make([]huh.Option[string], 0, len(options))
	for key, val := range options {
		out = append(out, huh.NewOption(key, val))
	}
	return out
}

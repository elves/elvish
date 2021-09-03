package main

import "src.elv.sh/pkg/cli/modes"

var items = []string{
	// Functions
	"length(", "read(", "scale(", "sqrt(",
	// Functions in math library
	"s(", "c(", "a(", "l(", "e(", "j(",
	// Statements
	"print ", "if ", "while (", "for (",
	"break", "continue", "halt", "return", "return (",
	// Pseudo statements
	"limits", "quit", "warranty",
}

func candidates() []modes.CompletionItem {
	candidates := make([]modes.CompletionItem, len(items))
	for i, item := range items {
		candidates[i] = modes.CompletionItem{ToShow: item, ToInsert: item}
	}
	return candidates
}

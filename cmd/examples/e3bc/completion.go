package main

import "src.elv.sh/pkg/cli/mode"

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

func candidates() []mode.CompletionItem {
	candidates := make([]mode.CompletionItem, len(items))
	for i, item := range items {
		candidates[i] = mode.CompletionItem{ToShow: item, ToInsert: item}
	}
	return candidates
}

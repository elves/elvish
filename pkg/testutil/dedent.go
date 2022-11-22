package testutil

import (
	"fmt"
	"strings"
)

// Dedent removes an optional leading newline, and removes the indentation
// present in the first line from all subsequent non-empty lines.
//
// Dedent panics if any non-empty line does not start with the same indentation
// as the first line.
func Dedent(text string) string {
	lines := strings.Split(strings.TrimPrefix(text, "\n"), "\n")
	line0 := lines[0]
	indent := line0[:len(line0)-len(strings.TrimLeft(lines[0], " \t"))]
	for i, line := range lines {
		if !strings.HasPrefix(line, indent) && line != "" {
			panic(fmt.Sprintf("line %d is not empty but doesn't start with %q", i, indent))
		}
		lines[i] = strings.TrimPrefix(line, indent)
	}
	return strings.Join(lines, "\n")
}

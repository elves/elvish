//go:build unix

package main

import (
	"fmt"
	"regexp"
	"strings"
)

type op struct {
	codeLines   []string
	tmuxCommand []string
}

var (
	ps1Pattern  = regexp.MustCompile(`^[~/][^ ]*> `)
	tmuxPattern = regexp.MustCompile(`^#[a-z]`)
)

func parseSpec(content string) ([]op, error) {
	lines := strings.Split(content, "\n")

	var ops []op
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		ps1 := ps1Pattern.FindString(line)
		if ps1 == "" {
			return nil, fmt.Errorf("invalid line %v", i+1)
		}
		content := line[len(ps1):]
		if tmuxPattern.MatchString(content) {
			ops = append(ops, op{tmuxCommand: strings.Fields(content[1:])})
			continue
		}

		codeLines := []string{content}
		ps2 := strings.Repeat(" ", len(ps1))
		for i++; i < len(lines) && strings.HasPrefix(lines[i], ps2); i++ {
			codeLines = append(codeLines, lines[i][len(ps2):])
		}
		i--
		ops = append(ops, op{codeLines: codeLines})
	}

	return ops, nil
}

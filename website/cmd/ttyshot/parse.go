//go:build unix

package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"src.elv.sh/pkg/transcript"
)

const (
	defaultRows = 100
	defaultCols = 52
)

type script struct {
	rows uint16
	cols uint16
	ops  []op
}

type op struct {
	code   string
	isTmux bool
}

var tmuxPattern = regexp.MustCompile(`^#[a-z]`)

func parseScript(name string, content []byte) (*script, error) {
	n, err := transcript.Parse(name, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	if len(n.Children) > 0 {
		return nil, fmt.Errorf("%s:%d: subsections not supported in ttyshot specs", name, n.Children[0].LineFrom)
	}
	s := script{rows: defaultRows, cols: defaultCols}
	for _, directive := range n.Directives {
		name, rest, _ := strings.Cut(directive, " ")
		switch name {
		case "rows":
			rows, err := strconv.ParseUint(rest, 0, 16)
			if err != nil {
				return nil, fmt.Errorf("parse rows %q: %w", rest, err)
			}
			s.rows = uint16(rows)
		case "cols":
			cols, err := strconv.ParseUint(rest, 0, 16)
			if err != nil {
				return nil, fmt.Errorf("parse cols %q: %w", rest, err)
			}
			s.cols = uint16(cols)
		default:
			return nil, fmt.Errorf("unknown directive %q in directive line %q", name, directive)
		}
	}
	for _, interaction := range n.Interactions {
		if interaction.Output != "" {
			return nil, fmt.Errorf("%s:%d: output not supported in ttyshot specs", name, interaction.OutputLineFrom)
		}
		if tmuxPattern.MatchString(interaction.Code) {
			s.ops = append(s.ops, op{code: interaction.Code[1:], isTmux: true})
		} else {
			s.ops = append(s.ops, op{code: interaction.Code})
		}
	}

	return &s, nil
}

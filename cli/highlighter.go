package cli

import (
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/styled"
)

// Highlighter represents a highlighter.
type Highlighter = clicore.Highlighter

// NewFuncHighlighter builds a Highlighter from a function that takes the code
// and returns styled text and a slice of errors.
func NewFuncHighlighter(f func(string) (styled.Text, []error)) Highlighter {
	return funcHighlighter{f}
}

// NewFuncHighlighterNoError builds a Highlighter from a function that takes the
// code and returns styled text.
func NewFuncHighlighterNoError(f func(string) styled.Text) Highlighter {
	return funcHighlighter{func(code string) (styled.Text, []error) {
		return f(code), nil
	}}
}

type funcHighlighter struct {
	f func(string) (styled.Text, []error)
}

func (hl funcHighlighter) Get(code string) (styled.Text, []error) {
	return hl.f(code)
}

func (hl funcHighlighter) LateUpdates() <-chan styled.Text {
	return nil
}

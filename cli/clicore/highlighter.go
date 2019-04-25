package clicore

import "github.com/elves/elvish/styled"

// Highlighter represents a code highlighter whose result can be delivered
// asynchronously.
type Highlighter interface {
	// Get returns the highlighted code and any static errors.
	Get(code string) (styled.Text, []error)
	// LateUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan styled.Text
}

func highlighterGet(hl Highlighter, code string) (styled.Text, []error) {
	if hl == nil {
		return styled.Plain(code), nil
	}
	return hl.Get(code)
}

// A Highlighter implementation useful for testing.
type fakeHighlighter struct {
	get         func(code string) (styled.Text, []error)
	lateUpdates chan styled.Text
}

func (hl fakeHighlighter) Get(code string) (styled.Text, []error) {
	return hl.get(code)
}

func (hl fakeHighlighter) LateUpdates() <-chan styled.Text {
	return hl.lateUpdates
}

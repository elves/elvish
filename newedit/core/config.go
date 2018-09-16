package core

import (
	"github.com/elves/elvish/styled"
)

type Config struct {
	RenderConfig   RenderConfig
	BeforeReadline []func()
	AfterReadline  []func(string)
}

type RenderConfig struct {
	MaxHeight   int
	Highlighter HighlighterCb
	Prompt      Prompt
	RPrompt     Prompt

	RPromptPersistent bool
}

type HighlighterCb func(string) (styled.Text, []error)

func (cb HighlighterCb) call(code string) (styled.Text, []error) {
	if cb == nil {
		return styled.Unstyled(code), nil
	}
	return cb(code)
}

// Prompt represents a prompt that can be delivered asynchronously.
type Prompt interface {
	// Trigger requests a re-computation of the prompt.
	Trigger()
	// Get returns the prompt.
	Get() styled.Text
	// LastUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan styled.Text
}

// A Prompt implementation that always return the same styled.Text.
type constPrompt struct{ t styled.Text }

func (constPrompt) Trigger()                        {}
func (p constPrompt) Get() styled.Text              { return p.t }
func (constPrompt) LateUpdates() <-chan styled.Text { return nil }

// Wraps a function into a Prompt.
type syncPrompt struct{ f func() styled.Text }

func (syncPrompt) Trigger()                        {}
func (p syncPrompt) Get() styled.Text              { return p.f() }
func (syncPrompt) LateUpdates() <-chan styled.Text { return nil }

package cli

import (
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/ui"
)

// AppSpec specifies the configuration and initial state for an App.
type AppSpec struct {
	TTY               TTY
	MaxHeight         func() int
	RPromptPersistent func() bool
	BeforeReadline    []func()
	AfterReadline     []func(string)

	Highlighter Highlighter
	Prompt      Prompt
	RPrompt     Prompt

	OverlayHandler el.Handler
	Abbreviations  func(f func(abbr, full string))
	QuotePaste     func() bool

	CodeAreaState codearea.State
	State         State
}

// Highlighter represents a code highlighter whose result can be delivered
// asynchronously.
type Highlighter interface {
	// Get returns the highlighted code and any static errors.
	Get(code string) (ui.Text, []error)
	// LateUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan ui.Text
}

// A Highlighter implementation that always returns plain text.
type dummyHighlighter struct{}

func (dummyHighlighter) Get(code string) (ui.Text, []error) {
	return ui.T(code), nil
}

func (dummyHighlighter) LateUpdates() <-chan ui.Text { return nil }

// Prompt represents a prompt whose result can be delivered asynchronously.
type Prompt interface {
	// Trigger requests a re-computation of the prompt. The force flag is set
	// when triggered for the first time during a ReadCode session or after a
	// SIGINT that resets the editor.
	Trigger(force bool)
	// Get returns the current prompt.
	Get() ui.Text
	// LastUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan ui.Text
}

// ConstPrompt is a Prompt implementation that always return the same ui.Text.
type ConstPrompt struct{ Content ui.Text }

func (ConstPrompt) Trigger(force bool)          {}
func (p ConstPrompt) Get() ui.Text              { return p.Content }
func (ConstPrompt) LateUpdates() <-chan ui.Text { return nil }

package cli

import (
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
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

	GlobalBindings   tk.Bindings
	CodeAreaBindings tk.Bindings
	QuotePaste       func() bool

	SimpleAbbreviations    func(f func(abbr, full string))
	CommandAbbreviations   func(f func(abbr, full string))
	SmallWordAbbreviations func(f func(abbr, full string))

	CodeAreaState tk.CodeAreaState
	State         State
}

// Highlighter represents a code highlighter whose result can be delivered
// asynchronously.
type Highlighter interface {
	// Get returns the highlighted code and any tips.
	Get(code string) (ui.Text, []ui.Text)
	// LateUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan struct{}
}

// A Highlighter implementation that always returns plain text.
type dummyHighlighter struct{}

func (dummyHighlighter) Get(code string) (ui.Text, []ui.Text) {
	return ui.T(code), nil
}

func (dummyHighlighter) LateUpdates() <-chan struct{} { return nil }

// Prompt represents a prompt whose result can be delivered asynchronously.
type Prompt interface {
	// Trigger requests a re-computation of the prompt. The force flag is set
	// when triggered for the first time during a ReadCode session or after a
	// SIGINT that resets the editor.
	Trigger(force bool)
	// Get returns the current prompt.
	Get() ui.Text
	// LastUpdates returns a channel for notifying late updates.
	LateUpdates() <-chan struct{}
}

// NewConstPrompt returns a Prompt that always shows the given text.
func NewConstPrompt(t ui.Text) Prompt {
	return constPrompt{t}
}

type constPrompt struct{ Content ui.Text }

func (constPrompt) Trigger(force bool)           {}
func (p constPrompt) Get() ui.Text               { return p.Content }
func (constPrompt) LateUpdates() <-chan struct{} { return nil }

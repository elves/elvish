package core

import (
	"sync"

	"github.com/elves/elvish/styled"
)

// Config wraps RawConfig for safe concurrent access. All of its methods are
// concurrency-safe, but direct field access must still be synchronized.
type Config struct {
	Raw   RawConfig
	Mutex sync.RWMutex
}

func (c *Config) TriggerPrompts(force bool) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	if c.Raw.Prompt != nil {
		c.Raw.Prompt.Trigger(force)
	}
	if c.Raw.RPrompt != nil {
		c.Raw.RPrompt.Trigger(force)
	}
}

// RawConfig keeps configurations of the editor.
type RawConfig struct {
	// Maximum lines of the terminal the editor may use. If MaxHeight <= 0,
	// there is no limit.
	MaxHeight int
	// Callback for highlighting the code the user has typed.
	Highlighter Highlighter
	// Left-hand prompt.
	Prompt Prompt
	// Right-hand prompt.
	RPrompt Prompt
	// Whether the rprompt is shown in the final redraw; in other words, whether
	// the rprompt persists in the terminal history when ReadCode returns.
	RPromptPersistent bool
}

// Highlighter is the type of callbacks for highlighting code.
type Highlighter func(string) (styled.Text, []error)

// Calls the highlighter, falling back to no highlighting if hl is nil.
func (hl Highlighter) call(code string) (styled.Text, []error) {
	if hl == nil {
		return styled.Unstyled(code), nil
	}
	return hl(code)
}

// Prompt represents a prompt whose result can be delivered asynchronously.
type Prompt interface {
	// Trigger requests a re-computation of the prompt. The force flag is set
	// when triggered for the first time during a ReadCode session or after a
	// SIGINT that resets the editor.
	Trigger(force bool)
	// Get returns the current prompt.
	Get() styled.Text
	// LastUpdates returns a channel for delivering late updates.
	LateUpdates() <-chan styled.Text
}

func promptGet(p Prompt) styled.Text {
	if p == nil {
		return nil
	}
	return p.Get()
}

// A Prompt implementation that always return the same styled.Text.
type constPrompt struct{ t styled.Text }

func (constPrompt) Trigger(force bool)              {}
func (p constPrompt) Get() styled.Text              { return p.t }
func (constPrompt) LateUpdates() <-chan styled.Text { return nil }

// Wraps a function into a Prompt.
type syncPrompt struct{ f func() styled.Text }

func (syncPrompt) Trigger(force bool)              {}
func (p syncPrompt) Get() styled.Text              { return p.f() }
func (syncPrompt) LateUpdates() <-chan styled.Text { return nil }

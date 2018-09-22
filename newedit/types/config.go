package types

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

// BeforeReadline returns c.Raw.BeforeReadline while r-locking c.Mutex.
func (c *Config) BeforeReadline() []func() {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.BeforeReadline
}

// AfterReadline returns c.Raw.AfterReadline while r-locking c.Mutex.
func (c *Config) AfterReadline() []func(string) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.AfterReadline
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
	// A list of functions called when ReadCode starts.
	BeforeReadline []func()
	// A list of functions called when ReadCode ends; the argument is the code
	// that has been read.
	AfterReadline []func(string)

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

// Call calls the highlighter. If hl is nil, it returns the unstyled text and no
// error.
func (hl Highlighter) Call(code string) (styled.Text, []error) {
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

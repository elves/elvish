package core

import (
	"sync"

	"github.com/elves/elvish/styled"
)

type Config struct {
	Raw   RawConfig
	Mutex sync.RWMutex
}

func (c *Config) BeforeReadline() []func() {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.BeforeReadline
}

func (c *Config) AfterReadline() []func(string) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.AfterReadline
}

func (c *Config) triggerPrompts() {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	if c.Raw.Prompt != nil {
		c.Raw.Prompt.Trigger()
	}
	if c.Raw.RPrompt != nil {
		c.Raw.RPrompt.Trigger()
	}
}

type RawConfig struct {
	BeforeReadline []func()
	AfterReadline  []func(string)

	MaxHeight         int
	Highlighter       HighlighterCb
	Prompt            Prompt
	RPrompt           Prompt
	RPromptPersistent bool
}

type renderSetup struct {
	height int
	width  int

	prompt  styled.Text
	rprompt styled.Text

	highlighter HighlighterCb
}

func makeRenderSetup(c *Config, h, w int) *renderSetup {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	if c.Raw.MaxHeight > 0 && c.Raw.MaxHeight < h {
		h = c.Raw.MaxHeight
	}
	return &renderSetup{
		h, w,
		promptGet(c.Raw.Prompt), promptGet(c.Raw.RPrompt), c.Raw.Highlighter}
}

func promptGet(p Prompt) styled.Text {
	if p == nil {
		return nil
	}
	return p.Get()
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

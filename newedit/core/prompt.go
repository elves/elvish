package core

import (
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/styled"
)

func promptGet(p types.Prompt) styled.Text {
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

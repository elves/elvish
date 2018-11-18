package core

import "github.com/elves/elvish/styled"

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

// A Prompt implementation useful for testing.
type fakePrompt struct {
	trigger     func(force bool)
	get         func() styled.Text
	lateUpdates chan styled.Text
}

func (p fakePrompt) Trigger(force bool) {
	if p.trigger != nil {
		p.trigger(force)
	}
}

func (p fakePrompt) Get() styled.Text {
	if p.get != nil {
		return p.get()
	}
	return nil
}

func (p fakePrompt) LateUpdates() <-chan styled.Text {
	return p.lateUpdates
}

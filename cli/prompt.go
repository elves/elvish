package cli

import (
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/styled"
)

// Prompt represents a prompt.
type Prompt = clicore.Prompt

// ConstPrompt builds a styled Prompt that does not change.
func ConstPrompt(t styled.Text) Prompt {
	return constPrompt{t}
}

// ConstPlainPrompt builds a plain Prompt that does not change.
func ConstPlainPrompt(s string) Prompt {
	return constPrompt{styled.Plain(s)}
}

// FuncPrompt builds a styled Prompt from a function.
func FuncPrompt(f func() styled.Text) Prompt {
	return funcPrompt{f}
}

// FuncPlainPrompt builds a plain Prompt from a function.
func FuncPlainPrompt(f func() string) Prompt {
	return funcPrompt{func() styled.Text { return styled.Plain(f()) }}
}

// A Prompt implementation that always return the same styled.Text.
type constPrompt struct{ t styled.Text }

func (constPrompt) Trigger(force bool)              {}
func (p constPrompt) Get() styled.Text              { return p.t }
func (constPrompt) LateUpdates() <-chan styled.Text { return nil }

type funcPrompt struct{ f func() styled.Text }

func (funcPrompt) Trigger(force bool)              {}
func (p funcPrompt) Get() styled.Text              { return p.f() }
func (funcPrompt) LateUpdates() <-chan styled.Text { return nil }

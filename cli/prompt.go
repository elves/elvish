package cli

import (
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/prompt"
	"github.com/elves/elvish/styled"
)

// Prompt represents a prompt.
type Prompt = clicore.Prompt

// NewConstPrompt builds a styled Prompt that does not change.
func NewConstPrompt(t styled.Text) Prompt {
	return constPrompt{t}
}

// NewConstPlainPrompt builds a plain Prompt that does not change.
func NewConstPlainPrompt(s string) Prompt {
	return constPrompt{styled.Plain(s)}
}

// NewFuncPrompt builds a styled Prompt from a function.
func NewFuncPrompt(f func() styled.Text) Prompt {
	return funcPrompt{f}
}

// NewFuncPlainPrompt builds a plain Prompt from a function.
func NewFuncPlainPrompt(f func() string) Prompt {
	return funcPrompt{func() styled.Text { return styled.Plain(f()) }}
}

// AsyncPromptConfig keeps configuration for async prompts.
type AsyncPromptConfig = prompt.RawConfig

// NewAsyncPrompt creates a Prompt that is updated asynchronously.
func NewAsyncPrompt(cfg *AsyncPromptConfig) Prompt {
	p := prompt.New(cfg.Compute)
	p.Config().Raw = *cfg
	return p
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

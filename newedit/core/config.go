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
	Prompt      PromptCb
	Rprompt     PromptCb

	RpromptPersistent bool
}

type HighlighterCb func(string) (styled.Text, []error)

func (cb HighlighterCb) call(code string) (styled.Text, []error) {
	if cb == nil {
		return styled.Text{styled.Segment{Text: code}}, nil
	}
	return cb(code)
}

type PromptCb func() styled.Text

func (cb PromptCb) call() styled.Text {
	if cb == nil {
		return nil
	}
	return cb()
}

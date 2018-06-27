package core

import (
	"github.com/elves/elvish/styled"
)

type Config struct {
	Render         *RenderConfig
	BeforeReadline []func()
	AfterReadline  []func(string)
}

func newConfig() *Config {
	return &Config{Render: &RenderConfig{
		Highlighter: dummyHighlighter,
		Prompt:      dummyPrompt,
		Rprompt:     dummyPrompt,
	}}
}

type RenderConfig struct {
	MaxHeight   int
	Highlighter HighlighterCb
	Prompt      PromptCb
	Rprompt     PromptCb

	RpromptPersistent bool
}

type HighlighterCb func(string) (styled.Text, []error)

func dummyHighlighter(s string) (styled.Text, []error) {
	return styled.Text{styled.Segment{Text: s}}, nil
}

type PromptCb func() styled.Text

func dummyPrompt() styled.Text {
	return nil
}

package clicore

import clitypes "github.com/elves/elvish/cli/clitypes"

// Config is the configuration for an App.
type Config interface {
	MaxHeight() int
	RPromptPersistent() bool
	BeforeReadline()
	AfterReadline(string)
	Highlighter() Highlighter
	Prompt() Prompt
	RPrompt() Prompt
	InitMode() clitypes.Mode
}

// DefaultConfig implements the Config interface, providing sensible default
// behavior. Other implementations of Config cam embed this struct and only
// implement the methods that it needs.
type DefaultConfig struct{}

// MaxHeight returns -1.
func (DefaultConfig) MaxHeight() int { return -1 }

// RPromptPersistent returns false.
func (DefaultConfig) RPromptPersistent() bool { return false }

// BeforeReadline does nothing.
func (DefaultConfig) BeforeReadline() {}

// AfterReadline does nothing.
func (DefaultConfig) AfterReadline(string) {}

// Highlighter returns nil.
func (DefaultConfig) Highlighter() Highlighter { return nil }

// Prompt returns nil.
func (DefaultConfig) Prompt() Prompt { return nil }

// RPrompt returns nil.
func (DefaultConfig) RPrompt() Prompt { return nil }

// InitMode returns nil.
func (DefaultConfig) InitMode() clitypes.Mode { return nil }

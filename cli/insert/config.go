package insert

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Config provides configuration for the insert mode.
type Config interface {
	// Handles the given key.
	HandleKey(ui.Key, *clitypes.State) clitypes.HandlerAction
	// Iterate all abbreviation pairs by calling the callback with each pair.
	IterateAbbr(func(abbr, full string))
	// Whether bracketed-pasted text should be quoted.
	QuotePaste() bool
}

// DefaultConfig implements the Config interface, providing sensible default
// behavior. Other implementations of Config can embed this struct and only
// implement the methods that it needs.
type DefaultConfig struct{}

// HandleKey calls cliutil.BasicHandler.
func (DefaultConfig) HandleKey(k ui.Key, st *clitypes.State) clitypes.HandlerAction {
	return cliutil.BasicHandler(term.KeyEvent(k), st)
}

// IterateAbbr is a no-op.
func (DefaultConfig) IterateAbbr(func(abbr, full string)) {}

// QuotePaste returns false.
func (DefaultConfig) QuotePaste() bool { return false }

package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/insert"
)

// InsertModeConfig is a struct containing configuration for the insert mode.
type InsertModeConfig struct {
	Binding    Binding
	Abbrs      StringPairs
	QuotePaste bool
}

// StringPairs is a general interface for accessing pairs of strings.
type StringPairs interface {
	IterateStringPairs(func(a, b string))
}

// NewSliceStringPairs builds a StringPairs from a slice.
func NewSliceStringPairs(s [][2]string) StringPairs {
	return sliceStringPairs(s)
}

type sliceStringPairs [][2]string

func (s sliceStringPairs) IterateStringPairs(f func(abbr, full string)) {
	for _, a := range s {
		f(a[0], a[1])
	}
}

func makeAbbrIterate(sp StringPairs) func(func(abbr, full string)) {
	if sp == nil {
		return nil
	}
	return sp.IterateStringPairs
}

// Initializes an insert mode.
func newInsertMode(cfg *InsertModeConfig, app *App) clitypes.Mode {
	return &insert.Mode{
		KeyHandler:  adaptBinding(cfg.Binding, app),
		AbbrIterate: makeAbbrIterate(cfg.Abbrs),
		Config: insert.Config{
			Raw: insert.RawConfig{
				QuotePaste: cfg.QuotePaste,
			},
		},
	}
}

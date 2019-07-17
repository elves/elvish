package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/insert"
	"github.com/elves/elvish/edit/ui"
)

// InsertModeConfig is a struct containing configuration for the insert mode.
type InsertModeConfig struct {
	Binding    Binding
	Abbrs      StringPairs
	QuotePaste bool
}

// Implements the insert.Config interface.
type insertModeConfig struct {
	*App
}

func (ic insertModeConfig) HandleKey(k ui.Key, st *clitypes.State) clitypes.HandlerAction {
	ic.cfg.Mutex.RLock()
	defer ic.cfg.Mutex.RUnlock()
	return handleKey(ic.cfg.InsertModeConfig.Binding, ic.App, k)
}

func (ic insertModeConfig) IterateAbbr(f func(abbr, full string)) {
	ic.cfg.Mutex.RLock()
	defer ic.cfg.Mutex.RUnlock()
	ic.cfg.InsertModeConfig.Abbrs.IterateStringPairs(f)
}

func (ic insertModeConfig) QuotePaste() bool {
	ic.cfg.Mutex.RLock()
	defer ic.cfg.Mutex.RUnlock()
	return ic.cfg.InsertModeConfig.QuotePaste
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
func newInsertMode(app *App) clitypes.Mode {
	return &insert.Mode{Config: insertModeConfig{app}}
}

// StartInsert starts the insert mode.
func StartInsert(ev KeyEvent) {
	ev.State().SetMode(ev.App().Insert)
}

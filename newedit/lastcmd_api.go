package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/lastcmd"
	"github.com/elves/elvish/newedit/listing"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(ed editor, ev *eval.Evaler, lsMode *listing.Mode) eval.Ns {
	m := lastcmd.Mode{Mode: lsMode}
	ns := eval.Ns{}.
		AddBuiltinFn("<edit:lastcmd>:", "start", func() {
			// TODO: Actually get the last line instead of using a fake one.
			m.Start("echo hello world", []string{"echo", "hello", "world"})
			ed.State().SetMode(m)
		})
	return ns
}

package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/lastcmd"
	"github.com/elves/elvish/newedit/listing"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(ed editor, ev *eval.Evaler, lsMode *listing.Mode) eval.Ns {
	ns := eval.Ns{}.
		AddBuiltinFn("edit:listing", "start", func() {
			lsMode.Start(lastcmd.StartConfig("echo hello world", []string{"echo", "hello", "world"}))
			ed.State().SetMode(lsMode)
		})
	return ns
}

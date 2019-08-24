package cliedit

import (
	"github.com/elves/elvish/cli/addons/lastcmd"
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(app *clicore.App, ev *eval.Evaler, lsBinding *bindingMap, store histutil.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:lastcmd>", "start", func() {
		// TODO: Specify wordifier
		lastcmd.Start(app, lastcmd.Config{binding, store, nil})
	})
}

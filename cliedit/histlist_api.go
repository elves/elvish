package cliedit

import (
	"github.com/elves/elvish/cli/addons/histlist"
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

// Initializes states for the histlist mode and its API.
func initHistlist(app *clicore.App, ev *eval.Evaler, lsBinding *bindingMap, store histutil.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:histlist>", "start", func() {
		histlist.Start(app, histlist.Config{binding, store})
	})
}

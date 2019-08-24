package cliedit

import (
	"github.com/elves/elvish/cli/addons/location"
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func initLocation(app *clicore.App, ev *eval.Evaler, lsBinding *bindingMap, store location.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:location>", "start", func() {
		location.Start(app, location.Config{binding, store})
	})
}

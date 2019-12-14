package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/stub"
	"github.com/elves/elvish/eval"
)

//elv:fn command:start
//
// Starts the command mode.

func initCommandAPI(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("command",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:command>:", map[string]interface{}{
			"start": func() {
				stub.Start(app, stub.Config{
					Binding: binding,
					Name:    " COMMAND ",
					Focus:   false,
				})
			},
		}))
}

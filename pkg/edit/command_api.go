package edit

import (
	"github.com/elves/elvish/pkg/cli/addons/stub"
	"github.com/elves/elvish/pkg/eval"
)

//elv:fn command:start
//
// Starts the command mode.

func initCommandAPI(ed *Editor, ev *eval.Evaler) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	ed.ns.AddNs("command",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:command>:", map[string]interface{}{
			"start": func() {
				stub.Start(ed.app, stub.Config{
					Binding: binding,
					Name:    " COMMAND ",
					Focus:   false,
				})
			},
		}))
}

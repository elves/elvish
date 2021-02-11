package edit

import (
	"src.elv.sh/pkg/cli/mode/stub"
	"src.elv.sh/pkg/eval"
)

//elv:fn command:start
//
// Starts the command mode.

func initCommandAPI(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	nb.AddNs("command",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:command>:", map[string]interface{}{
			"start": func() {
				stub.Start(ed.app, stub.Config{
					Binding: binding,
					Name:    " COMMAND ",
					Focus:   false,
				})
			},
		}).Ns())
}

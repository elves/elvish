package edit

// Implementation of the editor "command" mode.

import (
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/eval"
)

//elvdoc:var command:binding
//
// Key bindings for command mode. This is currently a very small subset of Vi
// command mode bindings.
//
// @cf edit:command:start

//elvdoc:fn command:start
//
// Enter command mode. This mode is intended to emulate Vi's command mode, but
// it is very incomplete right now.
//
// @cf edit:command:binding

func initCommandAPI(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	nb.AddNs("command",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:command>:", map[string]interface{}{
			"start": func() {
				w := mode.NewStub(mode.StubSpec{
					Bindings: bindings,
					Name:     " COMMAND ",
				})
				ed.app.PushAddon(w)
			},
		}).Ns())
}

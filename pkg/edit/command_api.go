package edit

import (
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/eval"
)

//elv:fn command:start
//
// Starts the command mode.

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
				ed.app.SetAddon(w, false)
			},
		}).Ns())
}

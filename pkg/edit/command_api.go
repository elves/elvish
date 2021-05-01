package edit

// Implementation of the editor "command" mode.

import (
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/eval"
)

//elvdoc:var command:binding
//
// Key bindings for command mode. The default bindings are a subset of Vi's command mode.
//
// TODO: Document the default bindings. For now note that they are codified in the
// *pkg/edit/default_bindings.go* source file. Specifically the `command:binding` assignment.
//
// @cf edit:command:start

//elvdoc:fn command:start
//
// Enter command mode. This is typically used to emulate the Vi editor's command mode by switching
// to the appropriate key bindings.
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
				ed.app.SetAddon(w, false)
			},
		}).Ns())
}

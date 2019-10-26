package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/histwalk"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
)

func initHistWalk(app *cli.App, ev *eval.Evaler, ns eval.Ns, fuser *histutil.Fuser) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("history",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:history>", map[string]interface{}{
			"start": func() {
				buf := app.CodeArea.CopyState().CodeBuffer
				walker := fuser.Walker(buf.Content[:buf.Dot])
				histwalk.Start(app, histwalk.Config{Binding: binding, Walker: walker})
			},
			"prev":  func() error { return histwalk.Prev(app) },
			"next":  func() error { return histwalk.Next(app) },
			"close": func() { histwalk.Close(app) },
		}))
}

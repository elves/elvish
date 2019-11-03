package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/histwalk"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
)

func initHistWalk(app cli.App, ev *eval.Evaler, ns eval.Ns, fuser *histutil.Fuser) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("history",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:history>", map[string]interface{}{
			"start":  func() { histWalkStart(app, fuser, binding) },
			"up":     func() error { return histwalk.Prev(app) },
			"down":   func() error { return histwalk.Next(app) },
			"accept": func() { histwalk.Accept(app) },
			"close":  func() { histwalk.Close(app) },
		}))
}

func histWalkStart(app cli.App, fuser *histutil.Fuser, binding el.Handler) {
	buf := app.CodeArea().CopyState().Buffer
	walker := fuser.Walker(buf.Content[:buf.Dot])
	histwalk.Start(app, histwalk.Config{Binding: binding, Walker: walker})
}

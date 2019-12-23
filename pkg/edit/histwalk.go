package edit

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/histwalk"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/eval"
)

func initHistWalk(app cli.App, ev *eval.Evaler, ns eval.Ns, fuser *histutil.Fuser) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("history",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:history>", map[string]interface{}{
			"start": func() { histWalkStart(app, fuser, binding) },
			"up":    func() { notifyIfError(app, histwalk.Prev(app)) },
			"down":  func() { notifyIfError(app, histwalk.Next(app)) },
			"down-or-quit": func() {
				err := histwalk.Next(app)
				if err == histutil.ErrEndOfHistory {
					histwalk.Close(app)
				} else {
					notifyIfError(app, err)
				}
			},
			"accept": func() { histwalk.Accept(app) },
			"close":  func() { histwalk.Close(app) },

			"fast-forward": fuser.FastForward,
		}))
}

func histWalkStart(app cli.App, fuser *histutil.Fuser, binding el.Handler) {
	buf := app.CodeArea().CopyState().Buffer
	walker := fuser.Walker(buf.Content[:buf.Dot])
	histwalk.Start(app, histwalk.Config{Binding: binding, Walker: walker})
}

func notifyIfError(app cli.App, err error) {
	if err != nil {
		app.Notify(err.Error())
	}
}

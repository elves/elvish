package edit

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/histwalk"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/eval"
)

//elvdoc:fn history:fast-forward
//
// Import command history entries that happened after the current session
// started.

func initHistWalk(ed *Editor, ev *eval.Evaler, hs *histStore) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	app := ed.app
	ed.ns.AddNs("history",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:history>", map[string]interface{}{
			"start": func() { histWalkStart(app, hs, binding) },
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

			"fast-forward": hs.FastForward,
		}))
}

func histWalkStart(app cli.App, hs *histStore, binding cli.Handler) {
	buf := app.CodeArea().CopyState().Buffer
	histwalk.Start(app, histwalk.Config{
		Binding: binding, Store: hs, Prefix: buf.Content[:buf.Dot]})
}

func notifyIfError(app cli.App, err error) {
	if err != nil {
		app.Notify(err.Error())
	}
}

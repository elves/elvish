package edit

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
)

//elvdoc:var history:binding
//
// Binding table for the history mode.

//elvdoc:fn history:start
//
// Starts the history mode.

//elvdoc:fn history:up
//
// Walks to the previous entry in history mode.

//elvdoc:fn history:down
//
// Walks to the next entry in history mode.

//elvdoc:fn history:down-or-quit
//
// Walks to the next entry in history mode, or quit the history mode if already
// at the newest entry.

//elvdoc:fn history:fast-forward
//
// Import command history entries that happened after the current session
// started.

func initHistWalk(ed *Editor, ev *eval.Evaler, hs *histStore, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	app := ed.app
	nb.AddNs("history",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:history>", map[string]interface{}{
			"start": func() { notifyError(app, histwalkStart(app, hs, bindings)) },
			"up":    func() { notifyError(app, histwalkDo(app, mode.Histwalk.Prev)) },

			"down": func() { notifyError(app, histwalkDo(app, mode.Histwalk.Next)) },
			"down-or-quit": func() {
				err := histwalkDo(app, mode.Histwalk.Next)
				if err == histutil.ErrEndOfHistory {
					app.SetAddon(nil, false)
				} else {
					notifyError(app, err)
				}
			},
			// TODO: Remove these builtins in favor of two universal accept and
			// close builtins
			"accept": func() { app.SetAddon(nil, true) },

			"fast-forward": hs.FastForward,
		}).Ns())
}

func histwalkStart(app cli.App, hs *histStore, bindings tk.Bindings) error {
	buf := app.CodeArea().CopyState().Buffer
	w, err := mode.NewHistwalk(app, mode.HistwalkSpec{
		Bindings: bindings, Store: hs, Prefix: buf.Content[:buf.Dot]})
	if w != nil {
		app.SetAddon(w, false)
	}
	return err
}

var errNotInHistoryMode = errors.New("not in history mode")

func histwalkDo(app cli.App, f func(mode.Histwalk) error) error {
	w, ok := app.CopyState().Addon.(mode.Histwalk)
	if !ok {
		return errNotInHistoryMode
	}
	return f(w)
}

func notifyError(app cli.App, err error) {
	if err != nil {
		app.Notify(err.Error())
	}
}

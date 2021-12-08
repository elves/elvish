package edit

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/modes"
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
		eval.BuildNsNamed("edit:history").
			AddVar("binding", bindingVar).
			AddGoFns(map[string]interface{}{
				"start": func() { notifyError(app, histwalkStart(app, hs, bindings)) },
				"up":    func() { notifyError(app, histwalkDo(app, modes.Histwalk.Prev)) },

				"down": func() { notifyError(app, histwalkDo(app, modes.Histwalk.Next)) },
				"down-or-quit": func() {
					err := histwalkDo(app, modes.Histwalk.Next)
					if err == histutil.ErrEndOfHistory {
						app.PopAddon()
					} else {
						notifyError(app, err)
					}
				},

				"fast-forward": hs.FastForward,
			}))
}

func histwalkStart(app cli.App, hs *histStore, bindings tk.Bindings) error {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return nil
	}
	buf := codeArea.CopyState().Buffer
	w, err := modes.NewHistwalk(app, modes.HistwalkSpec{
		Bindings: bindings, Store: hs, Prefix: buf.Content[:buf.Dot]})
	if w != nil {
		app.PushAddon(w)
	}
	return err
}

var errNotInHistoryMode = errors.New("not in history mode")

func histwalkDo(app cli.App, f func(modes.Histwalk) error) error {
	w, ok := app.ActiveWidget().(modes.Histwalk)
	if !ok {
		return errNotInHistoryMode
	}
	return f(w)
}

func notifyError(app cli.App, err error) {
	if err != nil {
		app.Notify(modes.ErrorText(err))
	}
}

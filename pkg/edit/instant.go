package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

func initInstant(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	nb.AddNs("-instant",
		eval.BuildNsNamed("edit:-instant").
			AddVar("binding", bindingVar).
			AddGoFns(map[string]any{
				"start": func() { instantStart(ed.app, ev, bindings) },
			}))
}

func instantStart(app cli.App, ev *eval.Evaler, bindings tk.Bindings) {
	execute := func(code string) ([]string, error) {
		outPort, collect, err := eval.StringCapturePort()
		if err != nil {
			return nil, err
		}
		ctx, done := eval.ListenInterrupts()
		err = ev.Eval(
			parse.Source{Name: "[instant]", Code: code},
			eval.EvalCfg{Ports: []*eval.Port{nil, outPort}, Interrupts: ctx})
		done()
		return collect(), err
	}
	w, err := modes.NewInstant(app,
		modes.InstantSpec{Bindings: bindings, Execute: execute})
	if w != nil {
		app.PushAddon(w)
		app.Redraw()
	}
	if err != nil {
		app.Notify(modes.ErrorText(err))
	}
}

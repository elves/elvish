package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

func initMinibuf(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	nb.AddNs("minibuf",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:minibuf>:", map[string]interface{}{
			"start": func() { minibufStart(ed, ev, binding) },
		}).Ns())
}

func minibufStart(ed *Editor, ev *eval.Evaler, binding tk.Handler) {
	w := tk.NewCodeArea(tk.CodeAreaSpec{
		Prompt:         mode.Prompt(" MINIBUF ", true),
		OverlayHandler: binding,
		OnSubmit:       func() { minibufSubmit(ed, ev) },
		// TODO: Add Highlighter. Right now the async highlighter is not
		// directly usable.
	})
	ed.app.MutateState(func(s *cli.State) { s.Addon = w })
	ed.app.Redraw()
}

func minibufSubmit(ed *Editor, ev *eval.Evaler) {
	app := ed.app
	codeArea, ok := app.CopyState().Addon.(tk.CodeArea)
	if !ok {
		return
	}
	ed.app.MutateState(func(s *cli.State) { s.Addon = nil })
	code := codeArea.CopyState().Buffer.Content
	src := parse.Source{Name: "[minibuf]", Code: code}
	notifyPort, cleanup := makeNotifyPort(ed)
	defer cleanup()
	ports := []*eval.Port{eval.DummyInputPort, notifyPort, notifyPort}
	err := ev.Eval(src, eval.EvalCfg{Ports: ports})
	if err != nil {
		app.Notify(err.Error())
	}
}

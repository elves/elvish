package edit

import (
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

func initMinibuf(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	nb.AddNs("minibuf",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:minibuf>:", map[string]interface{}{
			"start": func() { minibufStart(ed, ev, bindings) },
		}).Ns())
}

func minibufStart(ed *Editor, ev *eval.Evaler, bindings tk.Bindings) {
	w := tk.NewCodeArea(tk.CodeAreaSpec{
		Prompt:   mode.ModePrompt(" MINIBUF ", true),
		Bindings: bindings,
		OnSubmit: func() { minibufSubmit(ed, ev) },
		// TODO: Add Highlighter. Right now the async highlighter is not
		// directly usable.
	})
	ed.app.SetAddon(w, false)
	ed.app.Redraw()
}

func minibufSubmit(ed *Editor, ev *eval.Evaler) {
	app := ed.app
	codeArea, ok := app.CopyState().Addon.(tk.CodeArea)
	if !ok {
		return
	}
	ed.app.SetAddon(nil, false)
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

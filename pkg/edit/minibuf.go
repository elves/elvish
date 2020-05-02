package edit

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/parse"
)

func initMinibuf(ed *Editor, ev *eval.Evaler) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	ed.ns.AddNs("minibuf",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:minibuf>:", map[string]interface{}{
			"start": func() { minibufStart(ed.app, ev, binding) },
		}))
}

func minibufStart(app cli.App, ev *eval.Evaler, binding cli.Handler) {
	w := cli.NewCodeArea(cli.CodeAreaSpec{
		Prompt:         cli.ModePrompt(" MINIBUF ", true),
		OverlayHandler: binding,
		OnSubmit:       func() { minibufSubmit(app, ev) },
		// TODO: Add Highlighter. Right now the async highlighter is not
		// directly usable.
	})
	cli.SetAddon(app, w)
	app.Redraw()
}

func minibufSubmit(app cli.App, ev *eval.Evaler) {
	codeArea, ok := cli.Addon(app).(cli.CodeArea)
	if !ok {
		return
	}
	cli.SetAddon(app, nil)
	code := codeArea.CopyState().Buffer.Content
	src := parse.Source{Name: "[minibuf]", Code: code}
	op, err := ev.ParseAndCompile(src, nil)
	if err != nil {
		app.Notify(err.Error())
		return
	}
	notifyPort, cleanup := makeNotifyPort(app.Notify)
	defer cleanup()
	ports := []*eval.Port{eval.DevNullClosedChan, notifyPort, notifyPort}
	err = ev.Eval(op, eval.EvalCfg{Ports: ports})
	if err != nil {
		app.Notify(err.Error())
	}
}

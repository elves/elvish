package edit

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/el/layout"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/parse"
)

func initMinibuf(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("minibuf",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:minibuf>:", map[string]interface{}{
			"start": func() { minibufStart(app, ev, binding) },
		}))
}

func minibufStart(app cli.App, ev *eval.Evaler, binding el.Handler) {
	w := codearea.New(codearea.Spec{
		Prompt:         layout.ModePrompt(" MINIBUF ", true),
		OverlayHandler: binding,
		OnSubmit:       func() { minibufSubmit(app, ev) },
		// TODO: Add Highlighter. Right now the async highlighter is not
		// directly usable.
	})
	cli.SetAddon(app, w)
	app.Redraw()
}

func minibufSubmit(app cli.App, ev *eval.Evaler) {
	codeArea, ok := cli.Addon(app).(codearea.Widget)
	if !ok {
		return
	}
	cli.SetAddon(app, nil)
	code := codeArea.CopyState().Buffer.Content
	src := eval.NewInteractiveSource(code)
	n, err := parse.AsChunk("[minibuf]", code)
	if err != nil {
		app.Notify(err.Error())
		return
	}
	op, err := ev.Compile(n, src)
	if err != nil {
		app.Notify(err.Error())
		return
	}
	notifyPort, cleanup := makeNotifyPort(app.Notify)
	defer cleanup()
	ports := []*eval.Port{eval.DevNullClosedChan, notifyPort, notifyPort}
	err = ev.Eval(op, ports)
	if err != nil {
		app.Notify(err.Error())
	}
}

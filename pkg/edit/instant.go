package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/addons/instant"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

//elvdoc:var -instant:binding
//
// Binding for the instant mode.

//elvdoc:fn -instant:start
//
// Starts the instant mode. In instant mode, any text entered at the command
// line is evaluated immediately, with the output displayed.
//
// **WARNING**: Beware of unintended consequences when using destructive
// commands. For example, if you type `sudo rm -rf /tmp/*` in the instant mode,
// Elvish will attempt to evaluate `sudo rm -rf /` when you typed that far.

func initInstant(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	nb.AddNs("-instant",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:-instant>:", map[string]interface{}{
			"start": func() { instantStart(ed.app, ev, binding) },
		}).Ns())
}

func instantStart(app cli.App, ev *eval.Evaler, binding cli.Handler) {
	execute := func(code string) ([]string, error) {
		outPort, collect, err := eval.StringCapturePort()
		if err != nil {
			return nil, err
		}
		err = ev.Eval(
			parse.Source{Name: "[instant]", Code: code},
			eval.EvalCfg{
				Ports:     []*eval.Port{nil, outPort},
				Interrupt: eval.ListenInterrupts})
		return collect(), err
	}
	instant.Start(app, instant.Config{Binding: binding, Execute: execute})
}

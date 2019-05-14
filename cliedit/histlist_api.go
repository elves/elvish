package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
)

// Initializes states for the histlist mode and its API.
func initHistlist(ev *eval.Evaler, lsBinding *bindingMap, cfg *cli.HistlistModeConfig) eval.Ns {
	binding := emptyBindingMap
	cfg.Binding = newMapBinding(ev, &binding, lsBinding)
	ns := eval.Ns{}.
		AddGoFn("<edit:histlist>", "start", cli.StartHistlist)
	return ns
}

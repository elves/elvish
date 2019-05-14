package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(ev *eval.Evaler, lsBinding *bindingMap, cfg *cli.LastcmdModeConfig) eval.Ns {
	binding := emptyBindingMap
	cfg.Binding = newMapBinding(ev, &binding, lsBinding)
	ns := eval.Ns{}.
		AddGoFn("<edit:lastcmd>:", "start", cli.StartLastcmd)
	return ns
}

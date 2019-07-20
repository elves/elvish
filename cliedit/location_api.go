package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
)

func initLocation(ev *eval.Evaler, lsBinding *bindingMap, cfg *cli.LocationModeConfig) eval.Ns {
	binding := emptyBindingMap
	cfg.Binding = newMapBinding(ev, &binding, lsBinding)
	ns := eval.Ns{}.
		AddGoFn("<edit:location>", "start", cli.StartLocation)
	return ns
}

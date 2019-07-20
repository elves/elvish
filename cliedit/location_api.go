package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

func initLocation(ev *eval.Evaler, lsBinding *bindingMap, getDirs func() ([]storedefs.Dir, error), cd func(string) error, cfg *cli.LocationModeConfig) eval.Ns {
	binding := emptyBindingMap
	cfg.Binding = newMapBinding(ev, &binding, lsBinding)
	cfg.GetDirs = getDirs
	cfg.Cd = cd
	ns := eval.Ns{}.
		AddGoFn("<edit:location>", "start", cli.StartLocation)
	return ns
}

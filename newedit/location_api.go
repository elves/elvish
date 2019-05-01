package newedit

import (
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/location"
	"github.com/elves/elvish/store/storedefs"
)

func initLocation(a app, ev *eval.Evaler, getDirs func() ([]storedefs.Dir, error), cd func(string) error, lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := location.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(a, ev, &binding, lsBinding),
		Cd:         cd,
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:location>", "start", func() {
			startLocation(a, getDirs, &mode)
		})
	return ns
}

func startLocation(a app, getDirs func() ([]storedefs.Dir, error), mode *location.Mode) {
	dirs, err := getDirs()
	if err != nil {
		a.Notify("db error: " + err.Error())
	}
	mode.Start(dirs)
	a.State().SetMode(mode)
}

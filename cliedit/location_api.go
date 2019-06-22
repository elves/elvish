package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/cliedit/location"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

func initLocation(app *cli.App, ev *eval.Evaler, getDirs func() ([]storedefs.Dir, error), cd func(string) error, lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := location.Mode{
		Mode:       lsMode,
		KeyHandler: cli.AdaptBinding(newMapBinding(ev, &binding, lsBinding), app),
		Cd:         cd,
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:location>", "start", func(ev cli.KeyEvent) {
			startLocation(ev.App(), ev.State(), getDirs, &mode)
		})
	return ns
}

func startLocation(nt notifier, st *clitypes.State, getDirs func() ([]storedefs.Dir, error), mode *location.Mode) {
	dirs, err := getDirs()
	if err != nil {
		nt.Notify("db error: " + err.Error())
	}
	mode.Start(dirs)
	st.SetMode(mode)
}

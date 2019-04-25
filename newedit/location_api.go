package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/newedit/location"
	"github.com/elves/elvish/store/storedefs"
)

func initLocation(ed editor, ev *eval.Evaler, getDirs func() ([]storedefs.Dir, error), cd func(string) error, lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := location.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(ed, ev, &binding, lsBinding),
		Cd:         cd,
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:location>", "start", func() {
			startLocation(ed, getDirs, &mode)
		})
	return ns
}

func startLocation(ed editor, getDirs func() ([]storedefs.Dir, error), mode *location.Mode) {
	dirs, err := getDirs()
	if err != nil {
		ed.Notify("db error: " + err.Error())
	}
	mode.Start(dirs)
	ed.State().SetMode(mode)
}

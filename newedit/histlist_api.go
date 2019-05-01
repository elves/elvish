package newedit

import (
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/histlist"
	"github.com/elves/elvish/newedit/listing"
)

// Initializes states for the histlist mode and its API.
func initHistlist(a app, ev *eval.Evaler, getCmds func() ([]histutil.Entry, error), lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := histlist.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(a, ev, &binding, lsBinding),
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:histlist>", "start", func() {
			startHistlist(a, getCmds, &mode)
		})
	return ns
}

func startHistlist(a app, getCmds func() ([]histutil.Entry, error), mode *histlist.Mode) {
	cmds, err := getCmds()
	if err != nil {
		a.Notify("db error: " + err.Error())
	}
	mode.Start(cmds)
	a.State().SetMode(mode)
}

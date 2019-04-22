package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/histlist"
	"github.com/elves/elvish/newedit/listing"
)

// Initializes states for the histlist mode and its API.
func initHistlist(ed editor, ev *eval.Evaler, getCmds func() ([]string, error), lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := histlist.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(ed, ev, &binding, lsBinding),
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:histlist>", "start", func() {
			startHistlist(ed, getCmds, &mode)
		})
	return ns
}

func startHistlist(ed editor, getCmds func() ([]string, error), mode *histlist.Mode) {
	cmds, err := getCmds()
	if err != nil {
		ed.Notify("db error: " + err.Error())
	}
	mode.Start(cmds)
	ed.State().SetMode(mode)
}

package newedit

import (
	"github.com/elves/elvish/cli/lastcmd"
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/store/storedefs"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(a app, ev *eval.Evaler, st storedefs.Store, lsMode *listing.Mode, lsBinding *bindingMap) eval.Ns {
	binding := emptyBindingMap
	mode := lastcmd.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(a, ev, &binding, lsBinding),
	}
	ns := eval.Ns{}.
		AddGoFn("<edit:lastcmd>:", "start", func() {
			startLastcmd(a, st, &mode)
		})
	return ns
}

func startLastcmd(a app, st storedefs.Store, mode *lastcmd.Mode) {
	_, cmd, err := st.PrevCmd(-1, "")
	if err != nil {
		a.Notify("db error: " + err.Error())
		return
	}
	mode.Start(cmd, parseutil.Wordify(cmd))
	a.State().SetMode(mode)
}

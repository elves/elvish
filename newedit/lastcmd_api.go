package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/lastcmd"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/store/storedefs"
)

// Initializes states for the lastcmd mode and its API.
func initLastcmd(ed editor, ev *eval.Evaler, st storedefs.Store, lsMode *listing.Mode, lsBinding *BindingMap) eval.Ns {
	binding := EmptyBindingMap
	mode := lastcmd.Mode{
		Mode:       lsMode,
		KeyHandler: keyHandlerFromBindings(ed, ev, &binding, lsBinding),
	}
	ns := eval.Ns{}.
		AddBuiltinFn("<edit:lastcmd>:", "start", func() {
			startLastcmd(ed, st, &mode)
		})
	return ns
}

func startLastcmd(ed editor, st storedefs.Store, mode *lastcmd.Mode) {
	_, cmd, err := st.PrevCmd(-1, "")
	if err != nil {
		ed.Notify("db error: " + err.Error())
		return
	}
	mode.Start(cmd, parseutil.Wordify(cmd))
	ed.State().SetMode(mode)
}

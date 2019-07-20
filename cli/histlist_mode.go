package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histlist"
	"github.com/elves/elvish/edit/ui"
)

// HistlistModeConfig is a struct containing configuration for the histlist
// mode.
type HistlistModeConfig struct {
	Binding Binding
}

func newHistlist(app *App) *histlist.Mode {
	return &histlist.Mode{
		Mode: app.Listing,
		KeyHandler: func(k ui.Key) clitypes.HandlerAction {
			return handleKey(app.cfg.HistlistModeConfig.Binding, app, k)
		},
	}
}

// StartHistlist starts the history listing mode.
func StartHistlist(ev KeyEvent) {
	historyStore := ev.App().cfg.HistoryStore
	if historyStore == nil {
		ev.State().AddNote("no history store")
		return
	}
	cmds, err := historyStore.AllCmds()
	if err != nil {
		ev.State().AddNote("db error: " + err.Error())
		return
	}
	mode := ev.App().Histlist
	mode.Start(cmds)
	ev.State().SetMode(mode)
}

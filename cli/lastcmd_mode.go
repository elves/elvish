package cli

import (
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/lastcmd"
	"github.com/elves/elvish/edit/ui"
)

// LastcmdModeConfig is a struct containing configuration for the lastcmd mode.
type LastcmdModeConfig struct {
	Binding Binding
}

func newLastcmd(app *App) *lastcmd.Mode {
	return &lastcmd.Mode{
		Mode: app.Listing,
		KeyHandler: func(k ui.Key) clitypes.HandlerAction {
			return handleKey(app.cfg.LastcmdModeConfig.Binding, app, k)
		},
	}
}

// StartLastcmd starts the lastcmd mode.
func StartLastcmd(ev KeyEvent) {
	app := ev.App()
	cmd, err := app.cfg.HistoryStore.LastCmd()
	if err != nil {
		ev.State().AddNote("db error: " + err.Error())
		return
	}
	wordifier := app.cfg.Wordifier
	if wordifier == nil {
		wordifier = strings.Fields
	}
	words := wordifier(cmd.Text)
	app.Lastcmd.Start(cmd.Text, words)
	ev.State().SetMode(app.Lastcmd)
}

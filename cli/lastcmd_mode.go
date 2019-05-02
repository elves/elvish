package cli

import "strings"

// LastcmdModeConfig is a struct containing configuration for the lastcmd mode.
type LastcmdModeConfig struct {
	Binding Binding
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
	app.lastcmd.Start(cmd.Text, words)
	ev.State().SetMode(app.lastcmd)
}

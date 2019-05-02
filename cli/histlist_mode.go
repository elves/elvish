package cli

// HistlistModeConfig is a struct containing configuration for the histlist
// mode.
type HistlistModeConfig struct {
	Binding Binding
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
	mode := ev.App().histlist
	mode.Start(cmds)
	ev.State().SetMode(mode)
}

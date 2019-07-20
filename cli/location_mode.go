package cli

import (
	"os"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/location"
	"github.com/elves/elvish/edit/ui"
)

// LocationModeConfig is a struct containing configuration for the location
// mode.
type LocationModeConfig struct {
	Binding Binding
}

func getChdir(cfg *AppConfig) func(string) error {
	if cfg.DirStore == nil {
		return os.Chdir
	}
	return cfg.DirStore.Chdir
}

func newLocation(app *App) *location.Mode {
	return &location.Mode{
		Mode: app.Listing,
		KeyHandler: func(k ui.Key) clitypes.HandlerAction {
			return handleKey(app.cfg.LocationModeConfig.Binding, app, k)
		},
		Cd: getChdir(app.cfg),
	}
}

// StartLocation starts the location mode.
func StartLocation(ev KeyEvent) {
	app := ev.App()
	if app.cfg.DirStore == nil {
		app.Notify("no directory store")
		return
	}
	dirs, err := app.cfg.DirStore.Dirs()
	if err != nil {
		app.Notify("db error: " + err.Error())
		return
	}
	app.Location.Start(dirs)
	ev.State().SetMode(app.Location)
}

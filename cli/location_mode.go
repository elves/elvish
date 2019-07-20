package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/location"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
)

// LocationModeConfig is a struct containing configuration for the location
// mode.
type LocationModeConfig struct {
	Binding Binding
	Cd      func(string) error
	GetDirs func() ([]storedefs.Dir, error)
}

func newLocationMode(app *App) *location.Mode {
	return &location.Mode{
		Mode: app.Listing,
		KeyHandler: func(k ui.Key) clitypes.HandlerAction {
			return handleKey(app.cfg.LocationModeConfig.Binding, app, k)
		},
		Cd: app.cfg.LocationModeConfig.Cd,
	}
}

// StartLocation starts the location mode.
func StartLocation(ev KeyEvent) {
	app := ev.App()
	dirs, err := app.cfg.LocationModeConfig.GetDirs()
	if err != nil {
		app.Notify("db error: " + err.Error())
	}
	app.Location.Start(dirs)
	ev.State().SetMode(app.Location)
}

package cli

import (
	"github.com/elves/elvish/cli/clicore"
)

func DismissListing(app *clicore.App) {
	app.MutateAppState(func(s *clicore.State) { s.Listing = nil })
}

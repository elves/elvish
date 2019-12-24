package cli

import (
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
)

// CodeBuffer returns the code buffer of the main code area widget of the app.
func CodeBuffer(a App) codearea.CodeBuffer {
	return a.CodeArea().CopyState().Buffer
}

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(a App, buf codearea.CodeBuffer) {
	a.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
		s.Buffer = buf
	})
}

// Addon gets the current addon widget of the app.
func Addon(a App) el.Widget {
	return a.CopyState().Addon
}

// SetAddon sets the addon widget of the app.
func SetAddon(a App, addon el.Widget) {
	a.MutateState(func(s *State) { s.Addon = addon })
}

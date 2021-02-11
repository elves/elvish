package cli

import "src.elv.sh/pkg/cli/tk"

// CodeBuffer returns the code buffer of the main code area widget of the app.
func CodeBuffer(a App) tk.CodeBuffer {
	return a.CodeArea().CopyState().Buffer
}

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(a App, buf tk.CodeBuffer) {
	a.CodeArea().MutateState(func(s *tk.CodeAreaState) {
		s.Buffer = buf
	})
}

// Addon gets the current addon widget of the app.
func Addon(a App) tk.Widget {
	return a.CopyState().Addon
}

// SetAddon sets the addon widget of the app.
func SetAddon(a App, addon tk.Widget) {
	a.MutateState(func(s *State) { s.Addon = addon })
}

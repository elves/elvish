package cli

// GetCodeBuffer returns the code buffer of the main code area widget of the app.
func GetCodeBuffer(a App) CodeBuffer {
	return a.CodeArea().CopyState().Buffer
}

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(a App, buf CodeBuffer) {
	a.CodeArea().MutateState(func(s *CodeAreaState) {
		s.Buffer = buf
	})
}

// Addon gets the current addon widget of the app.
func Addon(a App) Widget {
	return a.CopyState().Addon
}

// SetAddon sets the addon widget of the app.
func SetAddon(a App, addon Widget) {
	a.MutateState(func(s *State) { s.Addon = addon })
}

package cli

import "github.com/elves/elvish/cli/el/codearea"

// CodeBuffer returns the code buffer of the main code area widget of the app.
func CodeBuffer(a App) codearea.Buffer {
	return a.CodeArea().CopyState().Buffer
}

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(a App, buf codearea.Buffer) {
	a.CodeArea().MutateState(func(s *codearea.State) {
		s.Buffer = buf
	})
}

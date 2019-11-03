// Package cliutil provides utilities based on the cli package.
package cliutil

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
)

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(app cli.App, buf codearea.CodeBuffer) {
	app.CodeArea().MutateState(func(s *codearea.State) {
		s.CodeBuffer = buf
	})
}

// GetCodeBuffer returns the code buffer of the main code area widget of the
// app.
func GetCodeBuffer(app cli.App) codearea.CodeBuffer {
	return app.CodeArea().CopyState().CodeBuffer
}

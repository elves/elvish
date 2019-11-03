// Package cliutil provides utilities based on the cli package.
package cliutil

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
)

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(app cli.App, buf codearea.Buffer) {
	app.CodeArea().MutateState(func(s *codearea.State) {
		s.Buffer = buf
	})
}

// GetCodeBuffer returns the code buffer of the main code area widget of the
// app.
func GetCodeBuffer(app cli.App) codearea.Buffer {
	return app.CodeArea().CopyState().Buffer
}

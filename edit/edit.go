// Package edit implements the Elvish command editor.
package edit

import (
	"os"

	"github.com/elves/elvish/edit/edcore"
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/eval"
)

// NewEditor creates an Editor. When the instance is no longer used, its Close
// method should be called.
func NewEditor(in *os.File, out *os.File, sigs <-chan os.Signal, ev *eval.Evaler) eddefs.Editor {
	return edcore.NewEditor(in, out, sigs, ev)
}

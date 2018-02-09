package eddefs

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Mode is an editor mode.
type Mode interface {
	ModeLine() ui.Renderer
	Binding(ui.Key) eval.Callable
	Teardown()
}

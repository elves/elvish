package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Mode is an editor mode.
type Mode interface {
	ModeLine() ui.Renderer
	Binding(*Editor, ui.Key) eval.Callable
}

type initer interface {
	Init()
}

type deiniter interface {
	Deinit()
}

// CursorOnModeLiner is an optional interface that modes can implement. If a
// mode does and the method returns true, the cursor is placed on the modeline
// when that mode is active.
type CursorOnModeLiner interface {
	CursorOnModeLine() bool
}

// Lister is a mode with a listing.
type Lister interface {
	List(maxHeight int) ui.Renderer
}

// ListRenderer is a mode with a listing that handles the rendering itself.
// NOTE(xiaq): This interface is being deprecated in favor of Lister.
type ListRenderer interface {
	// ListRender renders the listing under the given constraint of width and
	// maximum height. It returns a rendered buffer.
	ListRender(width, maxHeight int) *ui.Buffer
}

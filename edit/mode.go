package edit

import "github.com/elves/elvish/edit/ui"

// Additional interfaces mode implementations may satisfy.

type teardowner interface {
	Teardown()
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

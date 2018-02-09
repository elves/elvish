package edit

import "github.com/elves/elvish/edit/ui"

// Additional interfaces mode implementations may satisfy.

// cursorOnModeLiner is an optional interface that modes can implement. If a
// mode does and the method returns true, the cursor is placed on the modeline
// when that mode is active.
type cursorOnModeLiner interface {
	CursorOnModeLine() bool
}

// lister is an optional interface that modes can implement. If a mode
// implements this interface, the result of this method will be shown in the
// listing area.
type lister interface {
	List(maxHeight int) ui.Renderer
}

// listRenderer is similar to lister, but the mode handles the rendering itself.
// NOTE(xiaq): This interface is being deprecated in favor of Lister.
type listRenderer interface {
	// ListRender renders the listing under the given constraint of width and
	// maximum height. It returns a rendered buffer.
	ListRender(width, maxHeight int) *ui.Buffer
}

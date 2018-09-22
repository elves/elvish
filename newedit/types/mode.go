package types

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

// Mode is an editor mode; it handles keys and can affect the current UI.
type Mode interface {
	// Returns a Renderer for the modeline. If it returns nil, the modeline is
	// hidden.
	ModeLine() ui.Renderer
	// Returns a flag that can affect the UI.
	ModeRenderFlag() ModeRenderFlag
	// Handles a terminal event, and returns an action that can affect the editor
	// lifecycle.
	HandleEvent(tty.Event, *State) HandlerAction
}

// Lister is an optional interface that modes can implement. If a mode
// implements this interface, the result of this method is shown in the listing
// area.
type Lister interface {
	List(maxHeight int) ui.Renderer
}

// ModeRenderFlag is a bitmask for configuring the rendering behavior of modes.
type ModeRenderFlag uint

// Bits for ModeRenderFlag.
const (
	// Place the cursor on the mode line (instead of the code area).
	CursorOnModeLine = 1 << iota
	// Redraw the modeline after List. Has not effect if the mode does not
	// implement Lister.
	RedrawModeLineAfterList
)

// HandlerAction is used as the return code of Mode.HandleEvent and can affect
// the editor lifecycle.
type HandlerAction int

const (
	// NoAction is the default value of HandlerAction, which enacts no effect on
	// the editor lifecycle.
	NoAction HandlerAction = iota
	// CommitCode will cause the editor's ReadCode function to return with the
	// current code.
	CommitCode
	// CommitEOF
)

// A dummy Mode implementation.
type dummyMode struct{}

func (dummyMode) ModeLine() ui.Renderer {
	return nil
}

func (dummyMode) ModeRenderFlag() ModeRenderFlag {
	return 0
}

func (dummyMode) HandleEvent(_ tty.Event, _ *State) HandlerAction {
	return NoAction
}

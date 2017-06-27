package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Mode is an editor mode.
type Mode interface {
	ModeLine() renderer
	Binding(ui.Key) eval.CallableValue
}

type CursorOnModeLiner interface {
	CursorOnModeLine() bool
}

const (
	modeInsert         = "insert"
	modeRawInsert      = "raw-insert"
	modeCommand        = "command"
	modeCompletion     = "completion"
	modeNavigation     = "navigation"
	modeHistory        = "history"
	modeHistoryListing = "histlist"
	modeLastCmd        = "lastcmd"
	modeLocation       = "loc"
)

// ListRenderer is a mode with a listing.
type ListRenderer interface {
	// ListRender renders the listing under the given constraint of width and maximum
	// height. It returns a rendered buffer.
	ListRender(width, maxHeight int) *buffer
}

// Lister is a mode with a listing.
type Lister interface {
	List(maxHeight int) renderer
}

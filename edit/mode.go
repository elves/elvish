package edit

// Mode is an editor mode.
type Mode interface {
	Mode() ModeType
	ModeLine() renderer
}

type CursorOnModeLiner interface {
	CursorOnModeLine() bool
}

type ModeType int

const (
	modeInsert ModeType = iota
	modeCommand
	modeCompletion
	modeNavigation
	modeHistory
	modeHistoryListing
	modeBang
	modeLocation
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

package edit

// Mode is an editor mode.
type Mode interface {
	Mode() ModeType
	ModeLine() renderer
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

// OldLister is a mode with a listing.
type OldLister interface {
	// List renders the listing under the given constraint of width and maximum
	// height. It returns a rendered buffer.
	List(width, maxHeight int) *buffer
}

// Lister is a mode with a listing.
type Lister interface {
	List(maxHeight int) renderer
}

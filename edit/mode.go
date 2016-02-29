package edit

// Mode is an editor mode.
type Mode interface {
	Mode() ModeType
	// ModeLine renders a mode line under the given width constraint. It
	// returns a rendered buffer.
	ModeLine(width int) *buffer
}

type ModeType int

const (
	modeInsert ModeType = iota
	modeCommand
	modeCompletion
	modeNavigation
	modeHistory
	modeHistoryListing
	modeLocation
)

// Lister is a mode with a listing.
type Lister interface {
	// List renders the listing under the given constraint of width and maximum
	// height. It returns a rendered buffer.
	List(width, maxHeight int) *buffer
}

package edit

// Mode is an editor mode.
type Mode interface {
	Mode() ModeType
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

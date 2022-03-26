package term

import (
	"src.elv.sh/pkg/ui"
)

// Event represents an event that can be read from the terminal.
type Event interface {
	isEvent()
}

// KeyEvent represents a key press.
type KeyEvent ui.Key

// K constructs a new KeyEvent.
func K(r rune, mods ...ui.Mod) KeyEvent {
	return KeyEvent(ui.K(r, mods...))
}

// MouseEvent represents a mouse event (either pressing or releasing).
type MouseEvent struct {
	Pos
	Down bool
	// Number of the Button, 0-based. -1 for unknown.
	Button int
	Mod    ui.Mod
}

// CursorPosition represents a report of the current cursor position from the
// terminal driver, usually as a response from a cursor position request.
type CursorPosition Pos

// PasteSetting indicates the start or finish of pasted text.
type PasteSetting bool

// FatalErrorEvent represents an error that affects the Reader's ability to
// continue reading events. After sending a FatalError, the Reader makes no more
// attempts at continuing to read events and wait for Stop to be called.
type FatalErrorEvent struct{ Err error }

// NonfatalErrorEvent represents an error that can be gradually recovered. After
// sending a NonfatalError, the Reader will continue to read events. Note that
// one anamoly in the terminal might cause multiple NonfatalError events to be
// sent.
type NonfatalErrorEvent struct{ Err error }

func (KeyEvent) isEvent()   {}
func (MouseEvent) isEvent() {}

func (CursorPosition) isEvent() {}
func (PasteSetting) isEvent()   {}

func (FatalErrorEvent) isEvent()    {}
func (NonfatalErrorEvent) isEvent() {}

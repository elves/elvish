package tty

import "github.com/elves/elvish/edit/ui"

// Event represents an event that can be read from the terminal.
type Event interface {
	isEvent()
}

type RawRune rune
type KeyEvent ui.Key
type CursorPosition Pos
type PasteSetting bool

type MouseEvent struct {
	Pos
	Down bool
	// Number of the Button, 0-based. -1 for unknown.
	Button int
	Mod    ui.Mod
}

func (RawRune) isEvent()        {}
func (KeyEvent) isEvent()       {}
func (CursorPosition) isEvent() {}
func (MouseEvent) isEvent()     {}
func (PasteSetting) isEvent()   {}

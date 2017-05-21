package tty

import "github.com/elves/elvish/edit/uitypes"

// ReadUnit represents one "thing" that the Reader has read. It is one of the
// following: RawRune (when the reader is in the raw mode), Key, CursorPosition,
// MouseEvent, or PasteSetting.
type ReadUnit interface {
	isReadUnit()
}

type RawRune rune
type Key uitypes.Key
type CursorPosition Pos
type PasteSetting bool

func (RawRune) isReadUnit()        {}
func (Key) isReadUnit()            {}
func (CursorPosition) isReadUnit() {}
func (MouseEvent) isReadUnit()     {}
func (PasteSetting) isReadUnit()   {}

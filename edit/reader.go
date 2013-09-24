package edit

import (
	"fmt"
	"time"
	"../async"
)

var EscTimeout = time.Millisecond * 10

type Key struct {
	rune
	Ctrl bool
	Alt bool
}

var ZeroKey = Key{}

func PlainKey(r rune) Key {
	return Key{rune: r}
}

func CtrlKey(r rune) Key {
	return Key{rune: r, Ctrl: true}
}

func AltKey(r rune) Key {
	return Key{rune: r, Alt: true}
}

func (k Key) String() (s string) {
	if k.Ctrl {
		s += "Ctrl-"
	}
	if k.Alt {
		s += "Alt-"
	}
	if k.rune > 0 {
		s += string(k.rune)
	} else {
		s += fmt.Sprintf("(special %d)", k.rune)
	}
	return
}

const (
	F1 rune = -1-iota
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12

	Backspace // ^?

	Up // ^[OA
	Down // ^[OB
	Right // ^[OC
	Left // ^[OD

	Home // ^[[1~
	Insert // ^[[2~
	Delete // ^[[3~
	End // ^[[4~
	PageUp // ^[[5~
	PageDown // ^[[6~
)

// reader is the part of an Editor responsible for reading and decoding
// terminal key sequences.
type reader struct {
	runeReader *async.RuneReader
	readAhead []Key
}

func newReader(rr *async.RuneReader) *reader {
	return &reader{
		rr,
		make([]Key, 0),
	}
}

// type readerState func(rune) (bool, readerState)

func (rd *reader) readKey() (k Key, err error) {
	if n := len(rd.readAhead); n > 0 {
		k = rd.readAhead[0]
		rd.readAhead = rd.readAhead[1:]
		return
	}

	r, _, err := rd.runeReader.ReadRune()

	if err != nil {
		return
	}

	switch r {
	case 0x0:
		k = CtrlKey('`') // ^@
	case 0x1d:
		k = CtrlKey('6') // ^^
	case 0x1f:
		k = CtrlKey('/') // ^_
	case 0x7f: // ^? Backspace
		k = PlainKey(Backspace)
	case 0x1b: // ^[ Escape
		r, _, e := rd.runeReader.ReadRuneTimeout(EscTimeout)
		if e == async.Timeout {
			return CtrlKey('['), nil
		} else if e != nil {
			return ZeroKey, e
		}
		if r == '[' {
			r, _, e := rd.runeReader.ReadRuneTimeout(EscTimeout)
			if e == async.Timeout {
				return AltKey('['), nil
			} else if e != nil {
				return ZeroKey, e
			}
			switch r {
			case 'A':
				return PlainKey(Up), nil
			case 'B':
				return PlainKey(Down), nil
			case 'C':
				return PlainKey(Right), nil
			case 'D':
				return PlainKey(Left), nil
			default:
				rd.runeReader.UnreadRune()
				return AltKey('['), nil
			}
		}
		return AltKey(r), nil
	default:
		// Sane Ctrl- sequences that agree with the keyboard...
		if 0x1 <= r && r <= 0x1d {
			k = CtrlKey(r+0x40)
		} else {
			k = PlainKey(r)
		}
	}
	return
}

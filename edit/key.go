package edit

import (
	"fmt"
)

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

	Up
	Down
	Right
	Left

	Home
	Insert
	Delete
	End
	PageUp
	PageDown
)

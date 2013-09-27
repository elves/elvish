package edit

import (
	"fmt"
)

type Mod byte

const (
	Shift Mod = 1 << iota
	Alt
	Ctrl
)

type Key struct {
	rune
	Mod Mod
}

var ZeroKey = Key{}

func AltKey(r rune) Key {
	return Key{r, Alt}
}

func (k Key) String() (s string) {
	if k.Mod & Ctrl != 0 {
		s += "Ctrl-"
	}
	if k.Mod & Alt != 0 {
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

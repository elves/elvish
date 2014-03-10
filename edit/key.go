package edit

type Mod byte

const (
	Shift Mod = 1 << iota
	Alt
	Ctrl
)

type Key struct {
	Rune rune
	Mod  Mod
}

var ZeroKey = Key{}

func (k Key) String() (s string) {
	if k.Mod&Ctrl != 0 {
		s += "Ctrl-"
	}
	if k.Mod&Alt != 0 {
		s += "Alt-"
	}
	if k.Mod&Shift != 0 {
		s += "Shift-"
	}
	if k.Rune > 0 {
		if name, ok := KeyNames[k.Rune]; ok {
			s += name
		} else {
			s += string(k.Rune)
		}
	} else {
		s += FunctionKeyNames[-k.Rune]
	}
	return
}

const (
	Invalid rune = -iota
	F1
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
	DefaultBindingRune // Used in key of keyBinding for default binding

	// Some function key names are just aliases for their ASCII representation

	Tab       = '\t'
	Enter     = '\n'
	Backspace = 0x7f
)

var DefaultBinding = Key{DefaultBindingRune, 0}

var KeyNames = map[rune]string{
	Tab: "Tab", Enter: "Enter", Backspace: "Backspace",
}

var FunctionKeyNames = [...]string{
	"(Invalid)",
	"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
	"Up", "Down", "Right", "Left",
	"Home", "Insert", "Delete", "End", "PageUp", "PageDown",
}

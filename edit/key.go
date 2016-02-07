package edit

// Mod represents a modifier key.
type Mod byte

// Values for Mod.
const (
	// Shift is the shift modifier. It is only applied to special keys (e.g.
	// Shift-F1). For instance 'A' and '@' which are typically entered with the
	// shift key pressed, are not considered to be shift-modified.
	Shift Mod = 1 << iota
	// Alt is the alt modifier, traditionally known as the meta modifier.
	Alt
	Ctrl
)

// Key represents a single keyboard input, typically assembled from a escape
// sequence.
type Key struct {
	Rune rune
	Mod  Mod
}

// ZeroKey is the zero value of Key and is an invalid value.
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
		if name, ok := keyNames[k.Rune]; ok {
			s += name
		} else {
			s += string(k.Rune)
		}
	} else {
		s += functionKeyNames[-k.Rune]
	}
	return
}

// Special negative runes to represent function keys.
const (
	F1 rune = -iota - 1
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

	DefaultBindingRune // A special value used in DefaultBinding

	// Some function key names are just aliases for their ASCII representation

	Tab       = '\t'
	Enter     = '\n'
	Backspace = 0x7f
)

// DefaultBinding is an special value of Key, used as a key of keyBindings to
// indicate default binding.
var DefaultBinding = Key{DefaultBindingRune, 0}

var keyNames = map[rune]string{
	Tab: "Tab", Enter: "Enter", Backspace: "Backspace",
}

var functionKeyNames = [...]string{
	"(Invalid)",
	"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
	"Up", "Down", "Right", "Left",
	"Home", "Insert", "Delete", "End", "PageUp", "PageDown",
}

package edit

import (
	"fmt"
	"strings"
)

// Key represents a single keyboard input, typically assembled from a escape
// sequence.
type Key struct {
	Rune rune
	Mod  Mod
}

// Predefined special Key values.
var (
	// Default is used in the key binding table to indicate default binding.
	Default = Key{DefaultBindingRune, 0}
)

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
		i := int(-k.Rune)
		if i >= len(functionKeyNames) {
			s += fmt.Sprintf("(bad function key %d)", i)
		} else {
			s += functionKeyNames[-k.Rune]
		}
	}
	return
}

// modifierByName maps a name to an modifier. It is used for parsing keys where
// the modifier string is first turned to lower case, so that all of C, c,
// CTRL, Ctrl and ctrl can represent the Ctrl modifier.
var modifierByName = map[string]Mod{
	"s": Shift, "shift": Shift,
	"a": Alt, "alt": Alt,
	"m": Alt, "meta": Alt,
	"c": Ctrl, "ctrl": Ctrl,
}

// parseKey parses a key. The syntax is:
//
// Key = { Mod ('+' | '-') } BareKey
//
// BareKey = FunctionKeyName | SingleRune
func parseKey(s string) (Key, error) {
	var k Key
	// parse modifiers
	for {
		i := strings.IndexAny(s, "+-")
		if i == -1 {
			break
		}
		modname := strings.ToLower(s[:i])
		mod, ok := modifierByName[modname]
		if !ok {
			return Key{}, fmt.Errorf("bad modifier: %q", modname)
		}
		k.Mod |= mod
		s = s[i+1:]
	}

	if len(s) == 1 {
		k.Rune = rune(s[0])
		return k, nil
	}

	for r, name := range keyNames {
		if s == name {
			k.Rune = r
			return k, nil
		}
	}

	for i, name := range functionKeyNames[1:] {
		if s == name {
			k.Rune = rune(-i - 1)
			return k, nil
		}
	}

	return Key{}, fmt.Errorf("bad key: %q", s)
}

// Special negative runes to represent function keys, used in the Rune field of
// the Key struct.
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

var functionKeyNames = [...]string{
	"(Invalid)",
	"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
	"Up", "Down", "Right", "Left",
	"Home", "Insert", "Delete", "End", "PageUp", "PageDown", "default",
}

var keyNames = map[rune]string{
	Tab: "Tab", Enter: "Enter", Backspace: "Backspace",
}

package ui

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var ErrKeyMustBeString = errors.New("key must be string")

// Key represents a single keyboard input, typically assembled from a escape
// sequence.
type Key struct {
	Rune rune
	Mod  Mod
}

// Default is used in the key binding table to indicate default binding.
var Default = Key{DefaultBindingRune, 0}

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

	DefaultBindingRune // A special value to represent default binding.

	// Some function key names are just aliases for their ASCII representation

	Tab       = '\t'
	Enter     = '\n'
	Backspace = 0x7f
)

// functionKey stores the names of function keys, where the name of a function
// key k is stored at index -k. For instance, functionKeyNames[-F1] = "F1".
var functionKeyNames = [...]string{
	"(Invalid)",
	"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
	"Up", "Down", "Right", "Left",
	"Home", "Insert", "Delete", "End", "PageUp", "PageDown", "default",
}

// keyNames stores the name of function keys with a positive rune.
var keyNames = map[rune]string{
	Tab: "Tab", Enter: "Enter", Backspace: "Backspace",
}

func (k Key) String() string {
	var b bytes.Buffer
	if k.Mod&Ctrl != 0 {
		b.WriteString("Ctrl-")
	}
	if k.Mod&Alt != 0 {
		b.WriteString("Alt-")
	}
	if k.Mod&Shift != 0 {
		b.WriteString("Shift-")
	}
	if k.Rune > 0 {
		if name, ok := keyNames[k.Rune]; ok {
			b.WriteString(name)
		} else {
			b.WriteRune(k.Rune)
		}
	} else {
		i := int(-k.Rune)
		if i >= len(functionKeyNames) {
			fmt.Fprintf(&b, "(bad function key %d)", i)
		} else {
			b.WriteString(functionKeyNames[-k.Rune])
		}
	}
	return b.String()
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
		// XXX The following assumptions about keys with Ctrl are not checked
		// with all terminals.
		if k.Mod&Ctrl != 0 {
			// Keys with Ctrl as one of the modifiers and a single ASCII letter
			// as the base rune do not distinguish between cases. So we
			// normalize the base rune to upper case.
			if 'a' <= k.Rune && k.Rune <= 'z' {
				k.Rune += 'A' - 'a'
			}
			// Tab is equivalent to Ctrl-I and Ctrl-J is equivalent to Enter.
			// Normalize Ctrl-I to Tab and Ctrl-J to Enter.
			if k.Rune == 'I' {
				k.Mod &= ^Ctrl
				k.Rune = Tab
			} else if k.Rune == 'J' {
				k.Mod &= ^Ctrl
				k.Rune = Enter
			}
		}
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

// ToKey converts an elvish String to a Key. If the passed Value is not a
// String, it throws an error.
func ToKey(idx eval.Value) Key {
	skey, ok := idx.(eval.String)
	if !ok {
		util.Throw(ErrKeyMustBeString)
	}
	key, err := parseKey(string(skey))
	if err != nil {
		util.Throw(err)
	}
	return key
}

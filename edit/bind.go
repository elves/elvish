package edit

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval"
)

var keyBindings = map[bufferMode]map[Key]string{
	modeCommand: map[Key]string{
		Key{'i', 0}:    "start-insert",
		Key{'h', 0}:    "move-dot-left",
		Key{'l', 0}:    "move-dot-right",
		Key{'D', 0}:    "kill-line-right",
		DefaultBinding: "default-command",
	},
	modeInsert: map[Key]string{
		Key{'[', Ctrl}:    "start-command",
		Key{'U', Ctrl}:    "kill-line-left",
		Key{'K', Ctrl}:    "kill-line-right",
		Key{'W', Ctrl}:    "kill-word-left",
		Key{Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		Key{'H', Ctrl}:   "kill-rune-left",
		Key{Delete, 0}:   "kill-rune-right",
		Key{Left, 0}:     "move-dot-left",
		Key{Right, 0}:    "move-dot-right",
		Key{Left, Ctrl}:  "move-dot-left-word",
		Key{Right, Ctrl}: "move-dot-right-word",
		Key{Home, 0}:     "move-dot-sol",
		Key{End, 0}:      "move-dot-eol",
		Key{Up, Alt}:     "move-dot-up",
		Key{Down, Alt}:   "move-dot-down",
		Key{Enter, Alt}:  "insert-key",
		Key{Enter, 0}:    "return-line",
		Key{'D', Ctrl}:   "return-eof",
		Key{Tab, 0}:      "complete-prefix-or-start-completion",
		Key{Up, 0}:       "start-history",
		Key{'N', Ctrl}:   "start-navigation",
		DefaultBinding:   "default-insert",
	},
	modeCompletion: map[Key]string{
		Key{'[', Ctrl}: "cancel-completion",
		Key{Up, 0}:     "select-cand-up",
		Key{Down, 0}:   "select-cand-down",
		Key{Left, 0}:   "select-cand-left",
		Key{Right, 0}:  "select-cand-right",
		Key{Tab, 0}:    "cycle-cand-right",
		DefaultBinding: "default-completion",
	},
	modeNavigation: map[Key]string{
		Key{Up, 0}:     "select-nav-up",
		Key{Down, 0}:   "select-nav-down",
		Key{Left, 0}:   "ascend-nav",
		Key{Right, 0}:  "descend-nav",
		DefaultBinding: "default-navigation",
	},
	modeHistory: map[Key]string{
		Key{'[', Ctrl}: "start-insert",
		Key{Up, 0}:     "select-history-prev",
		Key{Down, 0}:   "select-history-next-or-quit",
		DefaultBinding: "default-history",
	},
}

var (
	invaliKey       = eval.NewFailure("invalid key to bind to")
	invalidFunction = eval.NewFailure("invalid function to bind")
)

var modifier = map[string]Mod{
	"s": Shift, "shift": Shift,
	"a": Alt, "alt": Alt,
	"m": Alt, "meta": Alt,
	"c": Ctrl, "ctrl": Ctrl,
}

func parseKey(s string) (Key, error) {
	var k Key
	// parse modifiers
	for {
		i := strings.IndexAny(s, "+-")
		if i == -1 {
			break
		}
		mod, ok := modifier[s[:i]]
		if !ok {
			return Key{}, fmt.Errorf("bad modifier: %q", mod)
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

// TODO Modify the binding table in ed instead of a global data structure.
func (ed *Editor) Bind(key string, function eval.Value) error {
	k, err := parseKey(key)
	if err != nil {
		return err
	}
	// TODO support functions
	s, ok := function.(eval.String)
	if !ok {
		return errors.New("function not string")
	}
	// TODO support other modes

	keyBindings[modeInsert][k] = string(s)

	return nil
}

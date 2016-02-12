package edit

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/eval"
)

var defaultBindings = map[bufferMode]map[Key]string{
	modeInsert: map[Key]string{
		DefaultBinding: "default-insert",
		// Moving.
		Key{Left, 0}:     "move-dot-left",
		Key{Right, 0}:    "move-dot-right",
		Key{Up, Alt}:     "move-dot-up",
		Key{Down, Alt}:   "move-dot-down",
		Key{Left, Ctrl}:  "move-dot-left-word",
		Key{Right, Ctrl}: "move-dot-right-word",
		Key{Home, 0}:     "move-dot-sol",
		Key{End, 0}:      "move-dot-eol",
		// Killing.
		Key{'U', Ctrl}:    "kill-line-left",
		Key{'K', Ctrl}:    "kill-line-right",
		Key{'W', Ctrl}:    "kill-word-left",
		Key{Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		Key{'H', Ctrl}: "kill-rune-left",
		Key{Delete, 0}: "kill-rune-right",
		// Inserting.
		Key{'.', Alt}:   "insert-last-word",
		Key{Enter, Alt}: "insert-key",
		// Controls.
		Key{Enter, 0}:  "return-line",
		Key{'D', Ctrl}: "return-eof",
		// Key{'[', Ctrl}: "startCommand",
		Key{Tab, 0}:    "complete-prefix-or-start-completion",
		Key{Up, 0}:     "start-history",
		Key{'N', Ctrl}: "start-navigation",
	},
	modeCommand: map[Key]string{
		DefaultBinding: "default-command",
		// Moving.
		Key{'h', 0}: "move-dot-left",
		Key{'l', 0}: "move-dot-right",
		Key{'k', 0}: "move-dot-up",
		Key{'j', 0}: "move-dot-down",
		Key{'b', 0}: "move-dot-left-word",
		Key{'w', 0}: "move-dot-right-word",
		Key{'0', 0}: "move-dot-sol",
		Key{'$', 0}: "move-dot-eol",
		// Killing.
		Key{'x', 0}: "kill-rune-right",
		Key{'D', 0}: "kill-line-right",
		// Controls.
		Key{'i', 0}: "start-insert",
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

var keyBindings = map[bufferMode]map[Key]Caller{}

var (
	errKeyMustBeString = errors.New("key must be string")
	errInvalidKey      = errors.New("invalid key to bind to")
	errInvalidFunction = errors.New("invalid function to bind")
)

// EvalCaller adapts an eval.Caller to a Caller.
type EvalCaller struct {
	eval.Caller
}

func (c EvalCaller) Repr() string {
	return c.Caller.Repr()
}

func (c EvalCaller) Call(ed *Editor) {
	// Input
	devnull, err := os.Open("/dev/null")
	if err != nil {
		Logger.Println(err)
		return
	}
	defer devnull.Close()
	in := make(chan eval.Value)
	close(in)

	// Output
	rout, out, err := os.Pipe()
	if err != nil {
		Logger.Println(err)
		return
	}
	chanOut := make(chan eval.Value)

	// Goroutines to collect output.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		rd := bufio.NewReader(rout)
		for {
			line, err := rd.ReadString('\n')
			Logger.Println("function writes bytes", line)
			if err != nil {
				break
			}
		}
		rout.Close()
		wg.Done()
	}()
	go func() {
		for v := range chanOut {
			Logger.Println("function writes Value", v.Repr())
		}
		wg.Done()
	}()

	ports := []*eval.Port{
		{File: devnull, Chan: in},
		{File: out, Chan: chanOut},
		{File: out, Chan: chanOut},
	}
	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor]", "", ports)
	ex := ec.PCall(c.Caller, []eval.Value{})
	if ex != nil {
		// XXX will disappear very quickly
		ed.pushTip("function error: " + ex.Error())
	}

	out.Close()
	close(chanOut)
	wg.Wait()
}

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
		modname := strings.ToLower(s[:i])
		mod, ok := modifier[modname]
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

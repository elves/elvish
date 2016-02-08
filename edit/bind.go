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

var keyBindings = map[bufferMode]map[Key]fn{
	modeCommand: map[Key]fn{
		Key{'i', 0}:    builtin(startInsert),
		Key{'h', 0}:    builtin(moveDotLeft),
		Key{'l', 0}:    builtin(moveDotRight),
		Key{'D', 0}:    builtin(killLineRight),
		DefaultBinding: builtin(defaultCommand),
	},
	modeInsert: map[Key]fn{
		Key{'[', Ctrl}:    builtin(startCommand),
		Key{'U', Ctrl}:    builtin(killLineLeft),
		Key{'K', Ctrl}:    builtin(killLineRight),
		Key{'W', Ctrl}:    builtin(killWordLeft),
		Key{Backspace, 0}: builtin(killRuneLeft),
		// Some terminal send ^H on backspace
		Key{'H', Ctrl}:   builtin(killRuneLeft),
		Key{Delete, 0}:   builtin(killRuneRight),
		Key{Left, 0}:     builtin(moveDotLeft),
		Key{Right, 0}:    builtin(moveDotRight),
		Key{Left, Ctrl}:  builtin(moveDotLeftWord),
		Key{Right, Ctrl}: builtin(moveDotRightWord),
		Key{Home, 0}:     builtin(moveDotSOL),
		Key{End, 0}:      builtin(moveDotEOL),
		Key{Up, Alt}:     builtin(moveDotUp),
		Key{Down, Alt}:   builtin(moveDotDown),
		Key{'.', Alt}:    builtin(insertLastWord),
		Key{Enter, Alt}:  builtin(insertKey),
		Key{Enter, 0}:    builtin(returnLine),
		Key{'D', Ctrl}:   builtin(returnEOF),
		Key{Tab, 0}:      builtin(completePrefixOrStartCompletion),
		Key{Up, 0}:       builtin(startHistory),
		Key{'N', Ctrl}:   builtin(startNavigation),
		DefaultBinding:   builtin(defaultInsert),
	},
	modeCompletion: map[Key]fn{
		Key{'[', Ctrl}: builtin(cancelCompletion),
		Key{Up, 0}:     builtin(selectCandUp),
		Key{Down, 0}:   builtin(selectCandDown),
		Key{Left, 0}:   builtin(selectCandLeft),
		Key{Right, 0}:  builtin(selectCandRight),
		Key{Tab, 0}:    builtin(cycleCandRight),
		DefaultBinding: builtin(defaultCompletion),
	},
	modeNavigation: map[Key]fn{
		Key{Up, 0}:     builtin(selectNavUp),
		Key{Down, 0}:   builtin(selectNavDown),
		Key{Left, 0}:   builtin(ascendNav),
		Key{Right, 0}:  builtin(descendNav),
		DefaultBinding: builtin(defaultNavigation),
	},
	modeHistory: map[Key]fn{
		Key{'[', Ctrl}: builtin(startInsert),
		Key{Up, 0}:     builtin(selectHistoryPrev),
		Key{Down, 0}:   builtin(selectHistoryNextOrQuit),
		DefaultBinding: builtin(defaultHistory),
	},
}

var (
	errInvalidKey      = errors.New("invalid key to bind to")
	errInvalidFunction = errors.New("invalid function to bind")
)

// CallerFn adapts an eval.Caller to an editor function.
type CallerFn struct {
	eval.Caller
}

func (c CallerFn) Call(ed *Editor) {
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

// Bind binds a key to a editor builtin or shell function.
func (ed *Editor) Bind(key string, function eval.Value) error {
	// TODO Modify the binding table in ed instead of a global data structure.
	k, err := parseKey(key)
	if err != nil {
		return err
	}

	var f fn
	switch function := function.(type) {
	case eval.String:
		builtin, ok := builtins[string(function)]
		if !ok {
			return fmt.Errorf("no builtin named %s", function.Repr())
		}
		f = builtin
	case eval.Caller:
		f = CallerFn{function}
	default:
		return fmt.Errorf("bad function type %s", function.Type())
	}

	keyBindings[modeInsert][k] = f

	return nil
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

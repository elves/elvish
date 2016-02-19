package edit

import (
	"bufio"
	"errors"
	"os"
	"sync"

	"github.com/elves/elvish/eval"
)

var defaultBindings = map[bufferMode]map[Key]string{
	modeInsert: map[Key]string{
		Default: "default-insert",
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
		Key{'H', Ctrl}: "start-history-listing",
	},
	modeCommand: map[Key]string{
		Default: "default-command",
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
		Key{Enter, 0}:  "accept-completion",
		Default:        "default-completion",
	},
	modeNavigation: map[Key]string{
		Key{Up, 0}:    "select-nav-up",
		Key{Down, 0}:  "select-nav-down",
		Key{Left, 0}:  "ascend-nav",
		Key{Right, 0}: "descend-nav",
		Default:       "default-navigation",
	},
	modeHistory: map[Key]string{
		Key{'[', Ctrl}: "start-insert",
		Key{Up, 0}:     "select-history-prev",
		Key{Down, 0}:   "select-history-next-or-quit",
		Default:        "default-history",
	},
	modeHistoryListing: map[Key]string{
		Default: "default-history-listing",
	},
}

var keyBindings = map[bufferMode]map[Key]Caller{}

var (
	errKeyMustBeString = errors.New("key must be string")
	errInvalidKey      = errors.New("invalid key to bind to")
	errInvalidFunction = errors.New("invalid function to bind")
)

// Caller is a function operating on an Editor. It is either a Builtin or an
// EvalCaller.
type Caller interface {
	eval.Reprer
	Call(ed *Editor)
}

func (b Builtin) Repr(int) string {
	return b.name
}

func (b Builtin) Call(ed *Editor) {
	b.impl(ed)
}

// EvalCaller adapts an eval.Caller to a Caller.
type EvalCaller struct {
	Caller eval.CallerValue
}

func (c EvalCaller) Repr(indent int) string {
	return c.Caller.Repr(indent)
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
			if err != nil {
				break
			}
			// XXX notify is not concurrency-safe.
			ed.notify("[bound fn bytes] %s", line[:len(line)-1])
		}
		rout.Close()
		wg.Done()
	}()
	go func() {
		for v := range chanOut {
			ed.notify("[bound fn value] %s", v.Repr(eval.NoPretty))
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
		ed.notify("function error: %s", ex.Error())
	}

	out.Close()
	close(chanOut)
	wg.Wait()

	ed.refresh(true, true)
}

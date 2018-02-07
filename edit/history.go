package edit

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vartypes"
)

// Command history mode.

type hist struct {
	ed      *Editor
	mutex   sync.RWMutex
	fuser   *history.Fuser
	binding BindingTable

	// Non-persistent state.
	walker *history.Walker
}

func init() {
	atEditorInit(initHist)
}

func initHist(ed *Editor, ns eval.Ns) {
	hist := &hist{ed: ed, binding: emptyBindingTable}
	ed.hist = hist

	if ed.daemon != nil {
		fuser, err := history.NewFuser(ed.daemon)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to initialize command history; disabled.")
		} else {
			hist.fuser = fuser
		}
	}

	subns := eval.Ns{
		"binding": eval.NewVariableFromPtr(&hist.binding),
		"list":    vartypes.NewRo(history.List{&hist.mutex, ed.daemon}),
	}
	subns.AddBuiltinFns("edit:history:", map[string]interface{}{
		"start":              hist.start,
		"up":                 hist.up,
		"down":               hist.down,
		"down-or-quit":       hist.downOrQuit,
		"switch-to-histlist": hist.switchToHistlist,
		"default":            hist.defaultFn,
	})

	ns.AddNs("history", subns)
}

func (h *hist) Binding(ed *Editor, k ui.Key) eval.Callable {
	return h.binding.getOrDefault(k)
}

func (h *hist) ModeLine() ui.Renderer {
	return modeLineRenderer{fmt.Sprintf(" HISTORY #%d ", h.walker.CurrentSeq()), ""}
}

func (hist *hist) start() {
	ed := hist.ed
	if hist.fuser == nil {
		ed.Notify("history offline")
		return
	}

	prefix := ed.buffer[:ed.dot]
	walker := hist.fuser.Walker(prefix)
	_, _, err := walker.Prev()

	if err == nil {
		hist.walker = walker
		ed.mode = hist
	} else {
		ed.addTip("no matching history item")
	}
}

var errNotHistory = errors.New("not in history mode")

func (hist *hist) up() {
	_, _, err := hist.walker.Prev()
	if err != nil {
		hist.ed.Notify("%s", err)
	}
}

func (hist *hist) down() {
	_, _, err := hist.walker.Next()
	if err != nil {
		hist.ed.Notify("%s", err)
	}
}

func (hist *hist) downOrQuit() {
	_, _, err := hist.walker.Next()
	if err != nil {
		hist.ed.mode = &hist.ed.insert
	}
}

func (hist *hist) switchToHistlist() {
	ed := hist.ed
	histlistStart(ed)
	if l, _, ok := getHistlist(ed); ok {
		ed.buffer = ""
		ed.dot = 0
		l.changeFilter(hist.walker.Prefix())
	}
}

func (hist *hist) defaultFn() {
	ed := hist.ed
	ed.buffer = hist.walker.CurrentCmd()
	ed.dot = len(ed.buffer)
	ed.mode = &ed.insert
	ed.setAction(reprocessKey)
}

func (ed *Editor) appendHistory(line string) {
	// TODO: should have a user variable to control the behavior
	// Do not add command leading by space into history. This is
	// useful for confidential operations.
	if strings.HasPrefix(line, " ") {
		return
	}

	if ed.daemon != nil && ed.hist.fuser != nil {
		ed.hist.mutex.Lock()
		go func() {
			err := ed.hist.fuser.AddCmd(line)
			ed.hist.mutex.Unlock()
			if err != nil {
				logger.Printf("Failed to AddCmd %q: %v", line, err)
			}
		}()
	}
}

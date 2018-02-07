package edit

import (
	"fmt"
	"os"
	"strings"
	"sync"

	. "github.com/elves/elvish/edit/edtypes"
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
	binding BindingMap

	// Non-persistent state.
	walker *history.Walker
}

func init() {
	atEditorInit(func(ed *Editor, ns eval.Ns) {
		ed.hist = initHist(ed, ns)
	})
}

func initHist(ed *Editor, ns eval.Ns) *hist {
	hist := &hist{ed: ed, binding: EmptyBindingMap}

	if ed.Daemon() != nil {
		fuser, err := history.NewFuser(ed.Daemon())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to initialize command history; disabled.")
		} else {
			hist.fuser = fuser
			ed.AddAfterReadline(hist.appendHistory)
		}
	}

	subns := eval.Ns{
		"binding": eval.NewVariableFromPtr(&hist.binding),
		"list":    vartypes.NewRo(history.List{&hist.mutex, ed.Daemon()}),
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

	return hist
}

func (h *hist) Binding(ed *Editor, k ui.Key) eval.Callable {
	return h.binding.GetOrDefault(k)
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

	buffer, dot := ed.Buffer()
	prefix := buffer[:dot]
	walker := hist.fuser.Walker(prefix)
	_, _, err := walker.Prev()

	if err == nil {
		hist.walker = walker
		ed.SetMode(hist)
	} else {
		ed.AddTip("no matching history item")
	}
}

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
		hist.ed.SetModeInsert()
	}
}

func (hist *hist) switchToHistlist() {
	ed := hist.ed
	histlistStart(ed)
	if l, _, ok := getHistlist(ed); ok {
		ed.SetBuffer("", 0)
		l.changeFilter(hist.walker.Prefix())
	}
}

func (hist *hist) defaultFn() {
	newBuffer := hist.walker.CurrentCmd()
	hist.ed.SetBuffer(newBuffer, len(newBuffer))
	hist.ed.SetModeInsert()
	hist.ed.SetAction(ReprocessKey)
}

func (hist *hist) appendHistory(line string) {
	// Do not add command leading by space into history. This is useful for
	// confidential operations.
	// TODO: Make this customizable.
	if strings.HasPrefix(line, " ") {
		return
	}

	if hist.fuser != nil {
		hist.mutex.Lock()
		go func() {
			err := hist.fuser.AddCmd(line)
			hist.mutex.Unlock()
			if err != nil {
				logger.Printf("Failed to AddCmd %q: %v", line, err)
			}
		}()
	}
}

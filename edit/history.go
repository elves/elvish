package edit

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
)

// Command history mode.

type hist struct {
	ed      eddefs.Editor
	mutex   sync.RWMutex
	fuser   *history.Fuser
	binding eddefs.BindingMap

	// Non-persistent state.
	walker *history.Walker
}

func init() {
	atEditorInit(initHist)
}

func initHist(ed *editor, ns eval.Ns) {
	hist := &hist{ed: ed, binding: emptyBindingMap}
	if ed.Daemon() != nil {
		fuser, err := history.NewFuser(ed.Daemon())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to initialize command history; disabled.")
		} else {
			hist.fuser = fuser
			ed.AddAfterReadline(hist.appendHistory)
		}
	}

	hl := &histlist{}
	histlistBinding := eddefs.EmptyBindingMap

	historyNs := eval.Ns{
		"binding": vars.NewFromPtr(&hist.binding),
		"list":    vars.NewRo(history.List{&hist.mutex, ed.Daemon()}),
	}
	historyNs.AddBuiltinFns("edit:history:", map[string]interface{}{
		"start":        hist.start,
		"up":           hist.up,
		"down":         hist.down,
		"down-or-quit": hist.downOrQuit,
		"default":      hist.defaultFn,
	})

	histlistNs := eval.Ns{
		"binding": vars.NewFromPtr(&histlistBinding),
	}
	histlistNs.AddBuiltinFns("edit:histlist:", map[string]interface{}{
		"start": func() {
			hl.start(ed, hist.fuser, histlistBinding)
		},
		"toggle-dedup":            func() { hl.toggleDedup(ed) },
		"toggle-case-sensitivity": func() { hl.toggleCaseSensitivity(ed) },
	})

	ns.AddNs("history", historyNs)
	ns.AddNs("histlist", histlistNs)
	// TODO(xiaq): Rename and put in edit:history
	ns.AddBuiltinFn("edit:", "command-history", hist.commandHistory)
}

func (h *hist) Teardown() {
	h.walker = nil
}

func (h *hist) Binding(k ui.Key) eval.Callable {
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

func (hist *hist) defaultFn() {
	newBuffer := hist.walker.CurrentCmd()
	hist.ed.SetBuffer(newBuffer, len(newBuffer))
	hist.ed.SetModeInsert()
	hist.ed.SetAction(reprocessKey)
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

func (hist *hist) commandHistory(fm *eval.Frame, args ...int) {
	var limit, start, end int

	out := fm.OutputChan()
	cmds, err := hist.fuser.AllCmds()
	if err != nil {
		return
	}

	if len(args) > 0 {
		limit = args[0]
	}

	total := len(cmds)
	switch {
	case limit > 0:
		start = 0
		end = limit
		if limit > total {
			end = total
		}
	case limit < 0:
		start = limit + total
		if start < 0 {
			start = 0
		}
		end = total
	default:
		start = 0
		end = total
	}

	for i := start; i < end; i++ {
		out <- vals.MakeMapFromKV(
			"id", strconv.Itoa(i),
			"cmd", cmds[i],
		)
	}
}

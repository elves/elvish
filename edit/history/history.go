package history

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[edit/history] ")

// Command history mode.

type hist struct {
	ed      eddefs.Editor
	mutex   sync.RWMutex
	fuser   *Fuser
	binding eddefs.BindingMap

	// Non-persistent state.
	walker    *Walker
	bufferLen int
}

func Init(ed eddefs.Editor, ns eval.Ns) {
	hist := &hist{ed: ed, binding: eddefs.EmptyBindingMap}
	if ed.Daemon() != nil {
		fuser, err := NewFuser(ed.Daemon())
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
		"binding": vars.FromPtr(&hist.binding),
		"list":    vars.NewRo(List{&hist.mutex, ed.Daemon()}),
	}
	historyNs.AddBuiltinFns("edit:history:", map[string]interface{}{
		"start":        hist.start,
		"up":           hist.up,
		"down":         hist.down,
		"down-or-quit": hist.downOrQuit,
		"default":      hist.defaultFn,

		"fast-forward": hist.fuser.FastForward,
	})

	histlistNs := eval.Ns{
		"binding": vars.FromPtr(&histlistBinding),
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
	h.bufferLen = 0
}

func (h *hist) Binding(k ui.Key) eval.Callable {
	return h.binding.GetOrDefault(k)
}

func (h *hist) ModeLine() ui.Renderer {
	return ui.NewModeLineRenderer(
		fmt.Sprintf(" HISTORY #%d ", h.walker.CurrentSeq()), "")
}

func (h *hist) Replacement() (int, int, string) {
	begin := len(h.walker.Prefix())
	return begin, h.bufferLen, h.walker.CurrentCmd()[begin:]
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
		hist.bufferLen = len(buffer)
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
	hist.ed.SetAction(eddefs.ReprocessKey)
}

func (hist *hist) appendHistory(line string) {
	// Do not add empty commands or commands with leading spaces to
	// TODO: Make this customizable.
	if line == "" || strings.HasPrefix(line, " ") {
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

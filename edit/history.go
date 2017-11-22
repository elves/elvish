package edit

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Command history mode.

var _ = registerBuiltins("history", map[string]func(*Editor){
	"start":              historyStart,
	"up":                 wrapHistoryBuiltin(historyUp),
	"down":               wrapHistoryBuiltin(historyDown),
	"down-or-quit":       wrapHistoryBuiltin(historyDownOrQuit),
	"switch-to-histlist": wrapHistoryBuiltin(historySwitchToHistlist),
	"default":            wrapHistoryBuiltin(historyDefault),
})

func init() {
	registerBindings(modeHistory, "history", map[ui.Key]string{
		{ui.Up, 0}:     "up",
		{ui.Down, 0}:   "down-or-quit",
		{'[', ui.Ctrl}: "insert:start",
		{'R', ui.Ctrl}: "switch-to-histlist",
		ui.Default:     "default",
	})
}

type hist struct {
	*history.Walker
}

func (*hist) Binding(m map[string]eval.Variable, k ui.Key) eval.CallableValue {
	return getBinding(m[modeHistory], k)
}

func (h *hist) ModeLine() ui.Renderer {
	return modeLineRenderer{fmt.Sprintf(" HISTORY #%d ", h.CurrentSeq()), ""}
}

func historyStart(ed *Editor) {
	if ed.historyFuser == nil {
		ed.Notify("history offline")
		return
	}
	prefix := ed.buffer[:ed.dot]
	walker := ed.historyFuser.Walker(prefix)
	hist := hist{walker}
	_, _, err := hist.Prev()
	if err == nil {
		ed.mode = &hist
	} else {
		ed.addTip("no matching history item")
	}
}

var errNotHistory = errors.New("not in history mode")

func wrapHistoryBuiltin(f func(*Editor, *hist)) func(*Editor) {
	return func(ed *Editor) {
		hist, ok := ed.mode.(*hist)
		if !ok {
			throw(errNotHistory)
		}
		f(ed, hist)
	}
}

func historyUp(ed *Editor, hist *hist) {
	_, _, err := hist.Prev()
	if err != nil {
		ed.Notify("%s", err)
	}
}

func historyDown(ed *Editor, hist *hist) {
	_, _, err := hist.Next()
	if err != nil {
		ed.Notify("%s", err)
	}
}

func historyDownOrQuit(ed *Editor, hist *hist) {
	_, _, err := hist.Next()
	if err != nil {
		ed.mode = &ed.insert
	}
}

func historySwitchToHistlist(ed *Editor, hist *hist) {
	histlistStart(ed)
	if hl := getHistlist(ed); hl != nil {
		ed.buffer = ""
		ed.dot = 0
		hl.changeFilter(hist.Prefix())
	}
}

func historyDefault(ed *Editor, hist *hist) {
	ed.buffer = hist.CurrentCmd()
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

	if ed.daemon != nil && ed.historyFuser != nil {
		ed.historyMutex.Lock()
		ed.daemon.Waits().Add(1)
		go func() {
			// TODO(xiaq): Report possible error
			err := ed.historyFuser.AddCmd(line)
			ed.daemon.Waits().Done()
			ed.historyMutex.Unlock()
			if err != nil {
				logger.Println("failed to add cmd to store:", err)
			} else {
				logger.Println("added cmd to store:", line)
			}
		}()
	}
}

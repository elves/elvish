package edit

import (
	"fmt"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store"
)

// Command history subsystem.

// Interface.

var _ = registerBuiltins("history", map[string]func(*Editor){
	"start":              historyStart,
	"up":                 historyUp,
	"down":               historyDown,
	"down-or-quit":       historyDownOrQuit,
	"switch-to-histlist": historySwitchToHistlist,
	"default":            historyDefault,
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
	current int
	prefix  string
	line    string
	// Maps content to the index of the last appearance. Used for deduplication.
	last map[string]int
}

func (hist) Mode() ModeType {
	return modeHistory
}

func (h *hist) ModeLine() renderer {
	return modeLineRenderer{fmt.Sprintf(" HISTORY #%d ", h.current), ""}
}

func historyStart(ed *Editor) {
	ed.hist.prefix = ed.line[:ed.dot]
	ed.hist.current = -1
	ed.hist.last = make(map[string]int)
	if ed.prevHistory() {
		ed.mode = &ed.hist
	} else {
		ed.addTip("no matching history item")
	}
}

func historyUp(ed *Editor) {
	ed.prevHistory()
}

func historyDown(ed *Editor) {
	ed.nextHistory()
}

func historyDownOrQuit(ed *Editor) {
	if !ed.nextHistory() {
		ed.mode = &ed.insert
	}
}

func historySwitchToHistlist(ed *Editor) {
	histlistStart(ed)
	if ed.mode == ed.histlist {
		ed.line = ""
		ed.dot = 0
		ed.histlist.changeFilter(ed.hist.prefix)
	}
}

func historyDefault(ed *Editor) {
	ed.acceptHistory()
	ed.mode = &ed.insert
	ed.nextAction = action{typ: reprocessKey}
}

// Implementation.

func (h *hist) jump(i int, line string) {
	h.current = i
	h.line = line
}

func (ed *Editor) appendHistory(line string) {
	if ed.store != nil {
		ed.historyMutex.Lock()
		ed.store.Waits.Add(1)
		go func() {
			// TODO(xiaq): Report possible error
			err := ed.store.AddCmd(line)
			ed.store.Waits.Done()
			ed.historyMutex.Unlock()
			if err != nil {
				logger.Println("failed to add cmd to store:", err)
			} else {
				logger.Println("added cmd to store:", line)
			}
		}()
	}
}

func (ed *Editor) prevHistory() bool {
	if ed.store == nil {
		return false
	}
	i := ed.hist.current
	var line string
	for {
		cmd, err := ed.store.GetLastCmd(i, ed.hist.prefix)
		i, line = cmd.Seq, cmd.Text
		if err != nil {
			if err != store.ErrNoMatchingCmd {
				logger.Println("LastCmd error:", err)
			}
			return false
		}
		if j, ok := ed.hist.last[line]; !ok || j == i {
			// Found the last among duplications
			ed.hist.last[line] = i
			break
		}
	}
	ed.hist.jump(i, line)
	return true
}

func (ed *Editor) nextHistory() bool {
	if ed.store == nil {
		return false
	}
	i := ed.hist.current
	var line string
	for {
		cmd, err := ed.store.GetFirstCmd(i+1, ed.hist.prefix)
		i, line = cmd.Seq, cmd.Text
		if err != nil {
			if err != store.ErrNoMatchingCmd {
				logger.Println("LastCmd error:", err)
			}
			return false
		}
		if ed.hist.last[line] == i {
			break
		}
	}

	ed.hist.jump(i, line)
	return true
}

// acceptHistory accepts the currently selected history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.hist.line
	ed.dot = len(ed.line)
}

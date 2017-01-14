package edit

import (
	"fmt"

	"github.com/elves/elvish/store"
)

// Command history subsystem.

// Interface.

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

func (h *hist) ModeLine(width int) *buffer {
	return makeModeLine(fmt.Sprintf(" HISTORY #%d ", h.current), width)
}

func startHistory(ed *Editor) {
	ed.ClearLastNotification()
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
	startHistlist(ed)
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
		go func() {
			ed.store.Waits.Add(1)
			// TODO(xiaq): Report possible error
			err := ed.store.AddCmd(line)
			ed.store.Waits.Done()
			ed.historyMutex.Unlock()
			if err != nil {
				Logger.Println("failed to add cmd to store:", err)
			} else {
				Logger.Println("added cmd to store:", line)
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
		var err error
		i, line, err = ed.store.LastCmd(i, ed.hist.prefix)
		if err != nil {
			if err != store.ErrNoMatchingCmd {
				Logger.Println("LastCmd error:", err)
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
		var err error
		i, line, err = ed.store.FirstCmd(i+1, ed.hist.prefix)
		if err != nil {
			if err != store.ErrNoMatchingCmd {
				Logger.Println("LastCmd error:", err)
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

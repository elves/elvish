package edit

import "fmt"

// Command history subsystem.

// Interface.

type historyState struct {
	current int
	prefix  string
	line    string
}

func (historyState) Mode() ModeType {
	return modeHistory
}

func (h *historyState) ModeLine(width int) *buffer {
	return makeModeLine(fmt.Sprintf("HISTORY #%d", h.current), width)
}

func startHistory(ed *Editor) {
	ed.history.prefix = ed.line[:ed.dot]
	ed.history.current = -1
	if ed.prevHistory() {
		ed.mode = &ed.history
	} else {
		ed.addTip("no matching history item")
	}
}

func selectHistoryPrev(ed *Editor) {
	ed.prevHistory()
}

func selectHistoryNext(ed *Editor) {
	ed.nextHistory()
}

func selectHistoryNextOrQuit(ed *Editor) {
	if !ed.nextHistory() {
		startInsert(ed)
	}
}

func defaultHistory(ed *Editor) {
	ed.acceptHistory()
	startInsert(ed)
	ed.nextAction = action{actionType: reprocessKey}
}

// Implementation.

func (h *historyState) jump(i int, line string) {
	h.current = i
	h.line = line
}

func (ed *Editor) appendHistory(line string) {
	if ed.store != nil {
		go func() {
			ed.store.Waits.Add(1)
			// TODO(xiaq): Report possible error
			ed.store.AddCmd(line)
			ed.store.Waits.Done()
			Logger.Println("added cmd to store:", line)
		}()
	}
}

func (ed *Editor) prevHistory() bool {
	if ed.store != nil {
		i, line, err := ed.store.LastCmd(ed.history.current, ed.history.prefix, true)
		if err == nil {
			ed.history.jump(i, line)
			return true
		}
		// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	}
	return false
}

func (ed *Editor) nextHistory() bool {
	if ed.store != nil {
		// Persistent history
		i, line, err := ed.store.FirstCmd(ed.history.current+1, ed.history.prefix, true)
		if err == nil {
			ed.history.jump(i, line)
			return true
		}
		// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	}

	return false
}

// acceptHistory accepts the currently selected history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.history.line
	ed.dot = len(ed.line)
}

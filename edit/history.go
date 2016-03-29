package edit

import "fmt"

// Command history subsystem.

// Interface.

type hist struct {
	current int
	prefix  string
	line    string
}

func (hist) Mode() ModeType {
	return modeHistory
}

func (h *hist) ModeLine(width int) *buffer {
	return makeModeLine(fmt.Sprintf(" HISTORY #%d ", h.current), width)
}

func startHistory(ed *Editor) {
	ed.hist.prefix = ed.line[:ed.dot]
	ed.hist.current = -1
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
		i, line, err := ed.store.LastCmd(ed.hist.current, ed.hist.prefix, true)
		if err == nil {
			ed.hist.jump(i, line)
			return true
		}
		// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	}
	return false
}

func (ed *Editor) nextHistory() bool {
	if ed.store != nil {
		// Persistent history
		i, line, err := ed.store.FirstCmd(ed.hist.current+1, ed.hist.prefix, true)
		if err == nil {
			ed.hist.jump(i, line)
			return true
		}
		// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	}

	return false
}

// acceptHistory accepts the currently selected history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.hist.line
	ed.dot = len(ed.line)
}

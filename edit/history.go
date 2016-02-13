package edit

// Command history subsystem.

type historyState struct {
	current int
	prefix  string
	line    string
}

func (h *historyState) jump(i int, line string) {
	h.current = i
	h.line = line
}

func (ed *Editor) appendHistory(line string) {
	if ed.store != nil {
		ed.store.AddCmd(line)
		// TODO(xiaq): Report possible error
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

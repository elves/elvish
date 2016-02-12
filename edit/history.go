package edit

import "strings"

// Command history subsystem.

type history struct {
	current int
	prefix  string
	line    string
}

func (h *history) jump(i int, line string) {
	h.current = i
	h.line = line
}

func (ed *Editor) appendHistory(line string) {
	ed.histories = append(ed.histories, line)
	if ed.store != nil {
		ed.store.AddCmd(line)
		// TODO(xiaq): Report possible error
	}
}

func lastHistory(histories []string, upto int, prefix string) (int, string) {
	for i := upto - 1; i >= 0; i-- {
		if strings.HasPrefix(histories[i], prefix) {
			return i, histories[i]
		}
	}
	return -1, ""
}

func firstHistory(histories []string, from int, prefix string) (int, string) {
	for i := from; i < len(histories); i++ {
		if strings.HasPrefix(histories[i], prefix) {
			return i, histories[i]
		}
	}
	return -1, ""
}

func (ed *Editor) prevHistory() bool {
	if ed.history.current > 0 {
		// Session history
		i, line := lastHistory(ed.histories, ed.history.current, ed.history.prefix)
		if i >= 0 {
			ed.history.jump(i, line)
			return true
		}
	}

	if ed.store != nil {
		// Persistent history
		upto := ed.cmdSeq + min(0, ed.history.current)
		i, line, err := ed.store.LastCmd(upto, ed.history.prefix, true)
		if err == nil {
			ed.history.jump(i-ed.cmdSeq, line)
			return true
		}
	}
	// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	return false
}

func (ed *Editor) nextHistory() bool {
	if ed.store != nil {
		// Persistent history
		if ed.history.current < -1 {
			from := ed.cmdSeq + ed.history.current + 1
			i, line, err := ed.store.FirstCmd(from, ed.history.prefix, true)
			if err == nil {
				ed.history.jump(i-ed.cmdSeq, line)
				return true
			}
			// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
		}
	}

	from := max(0, ed.history.current+1)
	i, line := firstHistory(ed.histories, from, ed.history.prefix)
	if i >= 0 {
		ed.history.jump(i, line)
		return true
	}
	return false
}

// acceptHistory accepts the currently selected history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.history.line
	ed.dot = len(ed.line)
}

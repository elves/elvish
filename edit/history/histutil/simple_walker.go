package histutil

import "strings"

// A simple implementation of Walker.
type simpleWalker struct {
	cmds   []string
	prefix string
	index  int
	// Index of the last occurrence of an element in cmds. Built on demand and
	// used for skipping duplicate entries.
	occ map[string]int
}

// NewSimpleWalker returns a Walker, given the slice of all commands and the prefix.
func NewSimpleWalker(cmds []string, prefix string) Walker {
	return &simpleWalker{cmds, prefix, len(cmds), make(map[string]int)}
}

func (w *simpleWalker) Prefix() string  { return w.prefix }
func (w *simpleWalker) CurrentSeq() int { return w.index }

func (w *simpleWalker) CurrentCmd() string {
	if w.index < len(w.cmds) {
		return w.cmds[w.index]
	}
	return ""
}

func (w *simpleWalker) Prev() error {
	for i := w.index - 1; i >= 0; i-- {
		cmd := w.cmds[i]
		if !strings.HasPrefix(cmd, w.prefix) {
			continue
		}
		j, ok := w.occ[cmd]
		if j == i || !ok {
			if !ok {
				w.occ[cmd] = i
			}
			w.index = i
			return nil
		}
	}
	return ErrEndOfHistory
}

func (w *simpleWalker) Next() error {
	if w.index >= len(w.cmds) {
		return ErrEndOfHistory
	}
	for w.index++; w.index < len(w.cmds); w.index++ {
		cmd := w.cmds[w.index]
		if !strings.HasPrefix(cmd, w.prefix) {
			continue
		}
		j, ok := w.occ[cmd]
		if ok && j == w.index {
			return nil
		}
	}
	return ErrEndOfHistory
}

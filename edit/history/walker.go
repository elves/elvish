package history

import (
	"errors"

	"github.com/elves/elvish/store"
)

var ErrEndOfHistory = errors.New("end of history")

// Walker is used for walking through history entries with a given (possibly
// empty) prefix, skipping duplicates entries.
type Walker struct {
	store  Store
	prefix string

	top     int // an index into the stack, the next element to return in Prev.
	stack   []string
	seq     []int
	inStack map[string]bool
}

func NewWalker(store Store, prefix string) *Walker {
	return &Walker{store, prefix, 0, nil, nil, map[string]bool{}}
}

// Prefix returns the prefix of the commands that the walker walks through.
func (w *Walker) Prefix() string {
	return w.prefix
}

// CurrentSeq returns the sequence number of the current entry.
func (w *Walker) CurrentSeq() int {
	if len(w.seq) > 0 && w.top <= len(w.seq) {
		return w.seq[w.top-1]
	}
	return -1
}

// CurrentSeq returns the content of the current entry.
func (w *Walker) CurrentCmd() string {
	if len(w.stack) > 0 && w.top <= len(w.stack) {
		return w.stack[w.top-1]
	}
	return ""
}

// Prev walks to the previous matching history entry, skipping all duplicates.
func (w *Walker) Prev() (int, string, error) {
	if w.top < len(w.stack) {
		w.top++
		return w.seq[w.top-1], w.stack[w.top-1], nil
	}
	seq := -1
	if len(w.seq) > 0 {
		seq = w.seq[len(w.seq)-1]
	}
	for {
		var (
			cmd string
			err error
		)
		seq, cmd, err = w.store.PrevCmd(seq, w.prefix)
		if err != nil {
			if err == store.ErrNoMatchingCmd {
				err = ErrEndOfHistory
			}
			return -1, "", err
		}
		if !w.inStack[cmd] {
			w.inStack[cmd] = true
			w.stack = append(w.stack, cmd)
			w.seq = append(w.seq, seq)
			w.top++
			return seq, cmd, nil
		}
	}
}

// Next reverses Prev.
func (w *Walker) Next() (int, string, error) {
	if w.top <= 0 {
		return -1, "", ErrEndOfHistory
	}
	w.top--
	if w.top == 0 {
		return -1, "", ErrEndOfHistory
	}
	return w.seq[w.top-1], w.stack[w.top-1], nil
}

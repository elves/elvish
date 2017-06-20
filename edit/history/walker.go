package history

import (
	"errors"
	"strings"

	"github.com/elves/elvish/store/storedefs"
)

var ErrEndOfHistory = errors.New("end of history")

// Walker is used for walking through history entries with a given (possibly
// empty) prefix, skipping duplicates entries.
type Walker struct {
	store       Store
	storeUpper  int
	sessionCmds []string
	sessionSeqs []int
	prefix      string

	// The next element to fetch from the session history. If equal to -1, the
	// next element comes from the storage backend.
	sessionIdx int
	// Index of the next element in the stack that Prev will return on next
	// call. If equal to len(stack), the next element needs to be fetched,
	// either from the session history or the storage backend.
	top     int
	stack   []string
	seq     []int
	inStack map[string]bool
}

func NewWalker(store Store, upper int, cmds []string, seqs []int, prefix string) *Walker {
	return &Walker{store, upper, cmds, seqs, prefix,
		len(cmds) - 1, 0, nil, nil, map[string]bool{}}
}

// Prefix returns the prefix of the commands that the walker walks through.
func (w *Walker) Prefix() string {
	return w.prefix
}

// CurrentSeq returns the sequence number of the current entry.
func (w *Walker) CurrentSeq() int {
	if len(w.seq) > 0 && w.top <= len(w.seq) && w.top > 0 {
		return w.seq[w.top-1]
	}
	return -1
}

// CurrentSeq returns the content of the current entry.
func (w *Walker) CurrentCmd() string {
	if len(w.stack) > 0 && w.top <= len(w.stack) && w.top > 0 {
		return w.stack[w.top-1]
	}
	return ""
}

// Prev walks to the previous matching history entry, skipping all duplicates.
func (w *Walker) Prev() (int, string, error) {
	// Entry comes from the stack.
	if w.top < len(w.stack) {
		i := w.top
		w.top++
		return w.seq[i], w.stack[i], nil
	}

	// Find the entry in the session part.
	for i := w.sessionIdx; i >= 0; i-- {
		seq := w.sessionSeqs[i]
		cmd := w.sessionCmds[i]
		if strings.HasPrefix(cmd, w.prefix) && !w.inStack[cmd] {
			w.push(cmd, seq)
			w.sessionIdx = i - 1
			return seq, cmd, nil
		}
	}
	// Not found in the session part.
	w.sessionIdx = -1

	seq := w.storeUpper
	if len(w.seq) > 0 && seq > w.seq[len(w.seq)-1] {
		seq = w.seq[len(w.seq)-1]
	}
	for {
		var (
			cmd string
			err error
		)
		seq, cmd, err = w.store.PrevCmd(seq, w.prefix)
		if err != nil {
			if err.Error() == storedefs.ErrNoMatchingCmd.Error() {
				err = ErrEndOfHistory
			}
			return -1, "", err
		}
		if !w.inStack[cmd] {
			w.push(cmd, seq)
			return seq, cmd, nil
		}
	}
}

func (w *Walker) push(cmd string, seq int) {
	w.inStack[cmd] = true
	w.stack = append(w.stack, cmd)
	w.seq = append(w.seq, seq)
	w.top++
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

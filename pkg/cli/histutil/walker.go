package histutil

import (
	"errors"
	"strings"

	"github.com/elves/elvish/pkg/store"
)

var ErrEndOfHistory = errors.New("end of history")

// Walker is used for walking through history entries with a given (possibly
// empty) prefix, skipping duplicates entries.
type Walker interface {
	Prefix() string
	CurrentSeq() int
	CurrentCmd() string
	Prev() error
	Next() error
}

type walker struct {
	store       DB
	storeUpper  int
	sessionCmds []store.Cmd
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

func NewWalker(store DB, upper int, cmds []store.Cmd, prefix string) Walker {
	return &walker{store, upper, cmds, prefix,
		len(cmds) - 1, 0, nil, nil, map[string]bool{}}
}

// Prefix returns the prefix of the commands that the walker walks through.
func (w *walker) Prefix() string {
	return w.prefix
}

// CurrentSeq returns the sequence number of the current entry.
func (w *walker) CurrentSeq() int {
	if len(w.seq) > 0 && w.top <= len(w.seq) && w.top > 0 {
		return w.seq[w.top-1]
	}
	return -1
}

// CurrentSeq returns the content of the current entry.
func (w *walker) CurrentCmd() string {
	if len(w.stack) > 0 && w.top <= len(w.stack) && w.top > 0 {
		return w.stack[w.top-1]
	}
	return ""
}

// Prev walks to the previous matching history entry, skipping all duplicates.
func (w *walker) Prev() error {
	// store.Cmd comes from the stack.
	if w.top < len(w.stack) {
		w.top++
		return nil
	}

	// Find the entry in the session part.
	for i := w.sessionIdx; i >= 0; i-- {
		cmd := w.sessionCmds[i]
		if strings.HasPrefix(cmd.Text, w.prefix) && !w.inStack[cmd.Text] {
			w.push(cmd.Text, cmd.Seq)
			w.sessionIdx = i - 1
			return nil
		}
	}
	// Not found in the session part.
	w.sessionIdx = -1

	seq := w.storeUpper
	if len(w.seq) > 0 && seq > w.seq[len(w.seq)-1] {
		seq = w.seq[len(w.seq)-1]
	}
	for {
		cmd, err := w.store.PrevCmd(seq, w.prefix)
		seq = cmd.Seq
		if err != nil {
			if err.Error() == store.ErrNoMatchingCmd.Error() {
				err = ErrEndOfHistory
			}
			return err
		}
		if !w.inStack[cmd.Text] {
			w.push(cmd.Text, seq)
			return nil
		}
	}
}

func (w *walker) push(cmd string, seq int) {
	w.inStack[cmd] = true
	w.stack = append(w.stack, cmd)
	w.seq = append(w.seq, seq)
	w.top++
}

// Next reverses Prev.
func (w *walker) Next() error {
	if w.top <= 1 {
		return ErrEndOfHistory
	}
	w.top--
	return nil
}

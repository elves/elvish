package history

import (
	"errors"
	"testing"

	"github.com/elves/elvish/store"
)

var (
	mockError = errors.New("mock error")
	mStore    = &mockStore{
		//              0       1        2         3        4         5
		cmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
	}
)

func TestWalker(t *testing.T) {
	var w *Walker

	wantCurrent := func(wantSeq int, wantCmd string) {
		seq, cmd := w.CurrentSeq(), w.CurrentCmd()
		if seq != wantSeq {
			t.Errorf("got seq %d, want %d", seq, wantSeq)
		}
		if cmd != wantCmd {
			t.Errorf("got cmd %q, want %q", cmd, wantCmd)
		}
	}
	wantCmd := func(f func() (int, string, error), wantSeq int, wantCmd string) {
		seq, cmd, err := f()
		if seq != wantSeq {
			t.Errorf("got seq %d, want %d", seq, wantSeq)
		}
		if cmd != wantCmd {
			t.Errorf("got cmd %q, want %q", cmd, wantCmd)
		}
		if err != nil {
			t.Errorf("got err %v, want nil", err)
		}
	}
	wantErr := func(f func() (int, string, error), want error) {
		_, _, err := f()
		if err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
	}

	w = NewWalker(mStore, "")
	wantCurrent(-1, "")
	wantCmd(w.Prev, 5, "ls a")
	wantCurrent(5, "ls a")
	wantErr(w.Next, ErrEndOfHistory)
	wantErr(w.Next, ErrEndOfHistory)
	wantCmd(w.Prev, 5, "ls a")

	wantCmd(w.Prev, 4, "echo a")
	wantCmd(w.Next, 5, "ls a")
	wantCmd(w.Prev, 4, "echo a")

	wantCmd(w.Prev, 3, "ls -a")
	// "echo a" should be skipped
	wantCmd(w.Prev, 1, "ls -l")
	wantCmd(w.Prev, 0, "echo")
	wantErr(w.Prev, ErrEndOfHistory)

	// Prefix matching 1.
	w = NewWalker(mStore, "echo")
	if w.Prefix() != "echo" {
		t.Errorf("got prefix %q, want %q", w.Prefix(), "echo")
	}
	wantCmd(w.Prev, 4, "echo a")
	wantCmd(w.Prev, 0, "echo")
	wantErr(w.Prev, ErrEndOfHistory)

	// Prefix matching 2.
	w = NewWalker(mStore, "ls")
	wantCmd(w.Prev, 5, "ls a")
	wantCmd(w.Prev, 3, "ls -a")
	wantCmd(w.Prev, 1, "ls -l")
	wantErr(w.Prev, ErrEndOfHistory)

	// Backend error.
	mStore.errAt = 3
	mStore.err = mockError
	w = NewWalker(mStore, "")
	wantCmd(w.Prev, 5, "ls a")
	wantCmd(w.Prev, 4, "echo a")
	wantErr(w.Prev, mockError)

	// store.ErrNoMatchingCmd is turned into ErrEndOfHistory.
	mStore.errAt = 5
	mStore.err = store.ErrNoMatchingCmd
	w = NewWalker(mStore, "")
	wantErr(w.Prev, ErrEndOfHistory)
}

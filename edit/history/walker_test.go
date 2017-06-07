package history

import (
	"errors"
	"testing"

	"github.com/elves/elvish/store"
)

func TestWalker(t *testing.T) {
	mockError := errors.New("mock error")
	walkerStore := &mockStore{
		//              0       1        2         3        4         5
		cmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
	}

	var w *Walker

	// Going back and forth.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	wantCurrent(t, w, -1, "")
	wantCmd(t, w.Prev, 5, "ls a")
	wantCurrent(t, w, 5, "ls a")
	wantErr(t, w.Next, ErrEndOfHistory)
	wantErr(t, w.Next, ErrEndOfHistory)
	wantCmd(t, w.Prev, 5, "ls a")

	wantCmd(t, w.Prev, 4, "echo a")
	wantCmd(t, w.Next, 5, "ls a")
	wantCmd(t, w.Prev, 4, "echo a")

	wantCmd(t, w.Prev, 3, "ls -a")
	// "echo a" should be skipped
	wantCmd(t, w.Prev, 1, "ls -l")
	wantCmd(t, w.Prev, 0, "echo")
	wantErr(t, w.Prev, ErrEndOfHistory)

	// With an upper bound on the storage.
	w = NewWalker(walkerStore, 2, nil, nil, "")
	wantCmd(t, w.Prev, 1, "ls -l")
	wantCmd(t, w.Prev, 0, "echo")
	wantErr(t, w.Prev, ErrEndOfHistory)

	// Prefix matching 1.
	w = NewWalker(walkerStore, -1, nil, nil, "echo")
	if w.Prefix() != "echo" {
		t.Errorf("got prefix %q, want %q", w.Prefix(), "echo")
	}
	wantCmd(t, w.Prev, 4, "echo a")
	wantCmd(t, w.Prev, 0, "echo")
	wantErr(t, w.Prev, ErrEndOfHistory)

	// Prefix matching 2.
	w = NewWalker(walkerStore, -1, nil, nil, "ls")
	wantCmd(t, w.Prev, 5, "ls a")
	wantCmd(t, w.Prev, 3, "ls -a")
	wantCmd(t, w.Prev, 1, "ls -l")
	wantErr(t, w.Prev, ErrEndOfHistory)

	// Walker with session history.
	w = NewWalker(walkerStore, -1,
		[]string{"ls -l", "ls -v", "echo haha"}, []int{7, 10, 12}, "ls")
	wantCmd(t, w.Prev, 10, "ls -v")

	wantCmd(t, w.Prev, 7, "ls -l")
	wantCmd(t, w.Next, 10, "ls -v")
	wantCmd(t, w.Prev, 7, "ls -l")

	wantCmd(t, w.Prev, 5, "ls a")
	wantCmd(t, w.Next, 7, "ls -l")
	wantCmd(t, w.Prev, 5, "ls a")

	wantCmd(t, w.Prev, 3, "ls -a")
	wantErr(t, w.Prev, ErrEndOfHistory)

	// Backend error.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	wantCmd(t, w.Prev, 5, "ls a")
	wantCmd(t, w.Prev, 4, "echo a")
	walkerStore.oneOffError = mockError
	wantErr(t, w.Prev, mockError)

	// store.ErrNoMatchingCmd is turned into ErrEndOfHistory.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	walkerStore.oneOffError = store.ErrNoMatchingCmd
	wantErr(t, w.Prev, ErrEndOfHistory)
}

func wantCurrent(t *testing.T, w *Walker, wantSeq int, wantCmd string) {
	seq, cmd := w.CurrentSeq(), w.CurrentCmd()
	if seq != wantSeq {
		t.Errorf("got seq %d, want %d", seq, wantSeq)
	}
	if cmd != wantCmd {
		t.Errorf("got cmd %q, want %q", cmd, wantCmd)
	}
}

func wantCmd(t *testing.T, f func() (int, string, error), wantSeq int, wantCmd string) {
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

func wantErr(t *testing.T, f func() (int, string, error), want error) {
	_, _, err := f()
	if err != want {
		t.Errorf("got err %v, want %v", err, want)
	}
}

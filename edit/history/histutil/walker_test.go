package histutil

import (
	"errors"
	"testing"

	"github.com/elves/elvish/store/storedefs"
)

func TestWalker(t *testing.T) {
	mockError := errors.New("mock error")
	walkerStore := &mockStore{
		//              0       1        2         3        4         5
		cmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
	}

	var w Walker
	wantCurrent := func(i int, s string) { checkWalkerCurrent(t, w, i, s) }
	wantErr := func(e, f error) { checkError(t, e, f) }

	// Going back and forth.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	wantCurrent(-1, "")
	w.Prev()
	wantCurrent(5, "ls a")
	wantCurrent(5, "ls a")
	wantErr(w.Next(), ErrEndOfHistory)
	wantErr(w.Next(), ErrEndOfHistory)
	w.Prev()
	wantCurrent(5, "ls a")

	w.Prev()
	wantCurrent(4, "echo a")
	w.Next()
	wantCurrent(5, "ls a")
	w.Prev()
	wantCurrent(4, "echo a")

	w.Prev()
	wantCurrent(3, "ls -a")
	// "echo a" should be skipped
	w.Prev()
	wantCurrent(1, "ls -l")
	w.Prev()
	wantCurrent(0, "echo")
	wantErr(w.Prev(), ErrEndOfHistory)

	// With an upper bound on the storage.
	w = NewWalker(walkerStore, 2, nil, nil, "")
	w.Prev()
	wantCurrent(1, "ls -l")
	w.Prev()
	wantCurrent(0, "echo")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Prefix matching 1.
	w = NewWalker(walkerStore, -1, nil, nil, "echo")
	if w.Prefix() != "echo" {
		t.Errorf("got prefix %q, want %q", w.Prefix(), "echo")
	}
	w.Prev()
	wantCurrent(4, "echo a")
	w.Prev()
	wantCurrent(0, "echo")
	w.Prev()
	wantErr(w.Prev(), ErrEndOfHistory)

	// Prefix matching 2.
	w = NewWalker(walkerStore, -1, nil, nil, "ls")
	w.Prev()
	wantCurrent(5, "ls a")
	w.Prev()
	wantCurrent(3, "ls -a")
	w.Prev()
	wantCurrent(1, "ls -l")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Walker with session history.
	w = NewWalker(walkerStore, -1,
		[]string{"ls -l", "ls -v", "echo haha"}, []int{7, 10, 12}, "ls")
	w.Prev()
	wantCurrent(10, "ls -v")

	w.Prev()
	wantCurrent(7, "ls -l")
	w.Next()
	wantCurrent(10, "ls -v")
	w.Prev()
	wantCurrent(7, "ls -l")

	w.Prev()
	wantCurrent(5, "ls a")
	w.Next()
	wantCurrent(7, "ls -l")
	w.Prev()
	wantCurrent(5, "ls a")

	w.Prev()
	wantCurrent(3, "ls -a")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Backend error.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	w.Prev()
	wantCurrent(5, "ls a")
	w.Prev()
	wantCurrent(4, "echo a")
	walkerStore.oneOffError = mockError
	wantErr(w.Prev(), mockError)

	// storedefs.ErrNoMatchingCmd is turned into ErrEndOfHistory.
	w = NewWalker(walkerStore, -1, nil, nil, "")
	walkerStore.oneOffError = storedefs.ErrNoMatchingCmd
	wantErr(w.Prev(), ErrEndOfHistory)
}

func checkWalkerCurrent(t *testing.T, w Walker, wantSeq int, wantCurrent string) {
	seq, cmd := w.CurrentSeq(), w.CurrentCmd()
	if seq != wantSeq {
		t.Errorf("got seq %d, want %d", seq, wantSeq)
	}
	if cmd != wantCurrent {
		t.Errorf("got cmd %q, want %q", cmd, wantCurrent)
	}
}

func checkError(t *testing.T, err, want error) {
	if err != want {
		t.Errorf("got err %v, want %v", err, want)
	}
}

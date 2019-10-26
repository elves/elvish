package histutil

import (
	"errors"
	"testing"

	"github.com/elves/elvish/store/storedefs"
)

func TestWalker(t *testing.T) {
	mockError := errors.New("mock error")
	walkerStore := &TestDB{
		//              0       1        2         3        4         5
		AllCmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
	}

	var w Walker
	wantCurrent := func(i int, s string) { t.Helper(); checkWalkerCurrent(t, w, i, s) }
	wantErr := func(e, f error) { t.Helper(); checkError(t, e, f) }
	wantOK := func(e error) { t.Helper(); checkError(t, e, nil) }

	// Going back and forth.
	w = NewWalker(walkerStore, -1, nil, "")
	wantCurrent(-1, "")
	wantOK(w.Prev())
	wantCurrent(5, "ls a")
	wantErr(w.Next(), ErrEndOfHistory)
	wantOK(w.Prev())
	wantCurrent(5, "ls a")

	wantOK(w.Prev())
	wantCurrent(4, "echo a")
	wantOK(w.Next())
	wantCurrent(5, "ls a")
	wantOK(w.Prev())
	wantCurrent(4, "echo a")

	wantOK(w.Prev())
	wantCurrent(3, "ls -a")
	// "echo a" should be skipped
	wantOK(w.Prev())
	wantCurrent(1, "ls -l")
	wantOK(w.Prev())
	wantCurrent(0, "echo")
	wantErr(w.Prev(), ErrEndOfHistory)

	// With an upper bound on the storage.
	w = NewWalker(walkerStore, 2, nil, "")
	wantOK(w.Prev())
	wantCurrent(1, "ls -l")
	wantOK(w.Prev())
	wantCurrent(0, "echo")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Prefix matching 1.
	w = NewWalker(walkerStore, -1, nil, "echo")
	if w.Prefix() != "echo" {
		t.Errorf("got prefix %q, want %q", w.Prefix(), "echo")
	}
	wantOK(w.Prev())
	wantCurrent(4, "echo a")
	wantOK(w.Prev())
	wantCurrent(0, "echo")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Prefix matching 2.
	w = NewWalker(walkerStore, -1, nil, "ls")
	wantOK(w.Prev())
	wantCurrent(5, "ls a")
	wantOK(w.Prev())
	wantCurrent(3, "ls -a")
	wantOK(w.Prev())
	wantCurrent(1, "ls -l")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Walker with session history.
	w = NewWalker(walkerStore, -1,
		[]Entry{{"ls -l", 7}, {"ls -v", 10}, {"echo haha", 12}}, "ls")
	wantOK(w.Prev())
	wantCurrent(10, "ls -v")

	wantOK(w.Prev())
	wantCurrent(7, "ls -l")
	wantOK(w.Next())
	wantCurrent(10, "ls -v")
	wantOK(w.Prev())
	wantCurrent(7, "ls -l")

	wantOK(w.Prev())
	wantCurrent(5, "ls a")
	wantOK(w.Next())
	wantCurrent(7, "ls -l")
	wantOK(w.Prev())
	wantCurrent(5, "ls a")

	wantOK(w.Prev())
	wantCurrent(3, "ls -a")
	wantErr(w.Prev(), ErrEndOfHistory)

	// Backend error.
	w = NewWalker(walkerStore, -1, nil, "")
	wantOK(w.Prev())
	wantCurrent(5, "ls a")
	wantOK(w.Prev())
	wantCurrent(4, "echo a")
	walkerStore.OneOffError = mockError
	wantErr(w.Prev(), mockError)

	// storedefs.ErrNoMatchingCmd is turned into ErrEndOfHistory.
	w = NewWalker(walkerStore, -1, nil, "")
	walkerStore.OneOffError = storedefs.ErrNoMatchingCmd
	wantErr(w.Prev(), ErrEndOfHistory)
}

func checkWalkerCurrent(t *testing.T, w Walker, wantSeq int, wantCurrent string) {
	t.Helper()
	seq, cmd := w.CurrentSeq(), w.CurrentCmd()
	if seq != wantSeq {
		t.Errorf("got seq %d, want %d", seq, wantSeq)
	}
	if cmd != wantCurrent {
		t.Errorf("got cmd %q, want %q", cmd, wantCurrent)
	}
}

func checkError(t *testing.T, err, want error) {
	t.Helper()
	if err != want {
		t.Errorf("got err %v, want %v", err, want)
	}
}

package histutil

import (
	"testing"

	"github.com/elves/elvish/pkg/store"
)

func TestDBStore_Cursor(t *testing.T) {
	db := NewFaultyInMemoryDB("+ 1", "- 2", "+ 3")
	s, err := NewDBStore(db)
	if err != nil {
		panic(err)
	}

	testCursorIteration(t, s.Cursor("+"), []store.Cmd{
		{Text: "+ 1", Seq: 0},
		{Text: "+ 3", Seq: 2},
	})

	// Test error conditions.
	c := s.Cursor("+")

	expect := func(wantCmd store.Cmd, wantErr error) {
		t.Helper()
		cmd, err := c.Get()
		if cmd != wantCmd {
			t.Errorf("Get -> %v, want %v", cmd, wantCmd)
		}
		if err != wantErr {
			t.Errorf("Get -> error %v, want %v", err, wantErr)
		}
	}

	db.SetOneOffError(mockError)
	c.Prev()
	expect(store.Cmd{Seq: 3}, mockError)

	c.Prev()
	expect(store.Cmd{Text: "+ 3", Seq: 2}, nil)

	db.SetOneOffError(mockError)
	c.Prev()
	expect(store.Cmd{Text: "+ 3", Seq: 2}, mockError)

	db.SetOneOffError(mockError)
	c.Next()
	expect(store.Cmd{Text: "+ 3", Seq: 2}, mockError)
}

// Remaing methods tested with HybridStore

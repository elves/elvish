package histutil

import (
	"testing"

	"src.elv.sh/pkg/store/storedefs"
)

func TestDBStore_Cursor(t *testing.T) {
	db := NewFaultyInMemoryDB("+ 1", "- 2", "+ 3")
	s, err := NewDBStore(db)
	if err != nil {
		panic(err)
	}

	testCursorIteration(t, s.Cursor("+"), []storedefs.Cmd{
		{Text: "+ 1", Seq: 0},
		{Text: "+ 3", Seq: 2},
	})

	// Test error conditions.
	c := s.Cursor("+")

	expect := func(wantCmd storedefs.Cmd, wantErr error) {
		t.Helper()
		cmd, err := c.Get()
		if cmd != wantCmd {
			t.Errorf("Get -> %v, want %v", cmd, wantCmd)
		}
		if err != wantErr {
			t.Errorf("Get -> error %v, want %v", err, wantErr)
		}
	}

	db.SetOneOffError(errMock)
	c.Prev()
	expect(storedefs.Cmd{Seq: 3}, errMock)

	c.Prev()
	expect(storedefs.Cmd{Text: "+ 3", Seq: 2}, nil)

	db.SetOneOffError(errMock)
	c.Prev()
	expect(storedefs.Cmd{Text: "+ 3", Seq: 2}, errMock)

	db.SetOneOffError(errMock)
	c.Next()
	expect(storedefs.Cmd{Text: "+ 3", Seq: 2}, errMock)
}

// Remaining methods tested with HybridStore.

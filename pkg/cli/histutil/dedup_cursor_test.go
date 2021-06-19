package histutil

import (
	"testing"

	"src.elv.sh/pkg/store/storedefs"
)

func TestDedupCursor(t *testing.T) {
	s := NewMemStore("0", "1", "2")
	c := NewDedupCursor(s.Cursor(""))

	wantCmds := []storedefs.Cmd{
		{Text: "0", Seq: 0},
		{Text: "1", Seq: 1},
		{Text: "2", Seq: 2}}

	testCursorIteration(t, c, wantCmds)
	// Go back again, this time with a full stack
	testCursorIteration(t, c, wantCmds)

	c = NewDedupCursor(s.Cursor(""))
	// Should be a no-op
	c.Next()
	testCursorIteration(t, c, wantCmds)

	c = NewDedupCursor(s.Cursor(""))
	c.Prev()
	c.Next()
	_, err := c.Get()
	if err != ErrEndOfHistory {
		t.Errorf("Get -> error %v, want ErrEndOfHistory", err)
	}
}

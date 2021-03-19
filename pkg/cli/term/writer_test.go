package term

import (
	"strings"
	"testing"
)

func TestWriter(t *testing.T) {
	sb := &strings.Builder{}
	testOutput := func(want string) {
		t.Helper()
		if sb.String() != want {
			t.Errorf("got %q, want %q", sb.String(), want)
		}
		sb.Reset()
	}

	w := NewWriter(sb)
	w.UpdateBuffer(
		NewBufferBuilder(10).Write("note 1").Buffer(),
		NewBufferBuilder(10).Write("line 1").SetDotHere().Buffer(),
		false)
	testOutput(hideCursor + "\rnote 1\033[K\n" + "line 1\r\033[6C" + showCursor)
}

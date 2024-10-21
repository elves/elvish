package term

import (
	"strings"
	"testing"

	"src.elv.sh/pkg/ui"
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
		ui.T("note 1"),
		NewBufferBuilder(10).Write("line 1").SetDotHere().Buffer(),
		false)
	testOutput(hideCursor + "\r \033[J\r\033[?7h\033[mnote 1\n\033[?7lline 1\r\033[6C" + showCursor)
}

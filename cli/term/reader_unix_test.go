// +build !windows,!plan9

package term

import (
	"os"
	"testing"
	"time"

	"github.com/elves/elvish/edit/ui"
)

func setupTest() (r, w *os.File, rd Reader) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w, NewReader(r)
}

var keyTests = []struct {
	input string
	want  Event
}{
	// Simple graphical key.
	{"x", K('x')},
	{"X", K('X')},
	{" ", K(' ')},

	// Ctrl key.
	{"\001", K('A', ui.Ctrl)},
	{"\033", K('[', ui.Ctrl)},

	// Ambiguous Ctrl keys; the reader uses the non-Ctrl form as canonical.
	{"\n", K('\n')},
	{"\t", K('\t')},
	{"\x7f", K('\x7f')}, // backspace

	// Alt plus simple graphical key.
	{"\033a", K('a', ui.Alt)},
	{"\033[", K('[', ui.Alt)},

	// G3-style key.
	{"\033OA", K(ui.Up)},
	{"\033OH", K(ui.Home)},

	// G3-style key with leading Escape.

	// CSI-sequence key identified by the ending rune.
	{"\033[A", K(ui.Up)},
	{"\033[H", K(ui.Home)},
	// Modifiers.
	{"\033[1;1A", K(ui.Up)},
	{"\033[1;2A", K(ui.Up, ui.Shift)},
	{"\033[1;3A", K(ui.Up, ui.Alt)},
	{"\033[1;4A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;5A", K(ui.Up, ui.Ctrl)},
	{"\033[1;6A", K(ui.Up, ui.Shift, ui.Ctrl)},
	{"\033[1;7A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;8A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},
	// The modifiers below should be for Meta, but we conflate Alt and Meta.
	{"\033[1;9A", K(ui.Up, ui.Alt)},
	{"\033[1;10A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;11A", K(ui.Up, ui.Alt)},
	{"\033[1;12A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;13A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;14A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},
	{"\033[1;15A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;16A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},

	// CSI-sequence key with one argument, always ending in '~'.
	{"\033[1~", K(ui.Home)},
	{"\033[11~", K(ui.F1)},

	// CSI-sequence key with three arguments and ending in '~'. The first
	// argument is always 27, the second identifies the modifier and the last
	// identifies the key.
	{"\033[27;4;63~", K(';', ui.Shift, ui.Alt)},
}

func TestReader_ReadKeys(t *testing.T) {
	r, w, reader := setupTest()
	defer r.Close()
	defer w.Close()
	reader.Start()
	defer reader.Close()
	defer reader.Stop()

	for _, test := range keyTests {
		w.WriteString(test.input)
		select {
		case event := <-reader.EventChan():
			if event != test.want {
				t.Errorf("Reader reads event %v, want %v", event, test.want)
			}
		case <-time.After(time.Second):
			t.Errorf("Reader timed out")
		}
	}
}

func TestReader_StopMakesUnderlyingFileAvailable(t *testing.T) {
	r, w, reader := setupTest()
	defer r.Close()
	defer w.Close()
	reader.Start()
	defer reader.Close()

	// tests that after calling Stop, the
	// Reader no longer attempts to read from the underlying file, so it is
	// available for use by others.
	reader.Stop()

	// Verify that the reader has indeed stopped: write something via w,
	// and get it back via r.
	written := "lorem ipsum"
	w.WriteString(written)
	var buf [32]byte
	nr, err := r.Read(buf[:])
	if err != nil {
		panic(err)
	}
	got := string(buf[:nr])
	if got != written {
		t.Errorf("got %q, want %q", got, written)
	}
}

func TestReader_StartAfterStopIndeedStarts(t *testing.T) {
	r, w, reader := setupTest()
	defer r.Close()
	defer w.Close()
	reader.Start()
	defer reader.Close()
	defer reader.Stop()

	for i := 0; i < 100; i++ {
		// Test that calling Start very shortly after Stop puts the Reader
		// in the correct started state. This test is for ensuring that the
		// operations do not have race conditions.
		reader.Stop()
		reader.Start()

		w.WriteString("a")
		select {
		case event := <-reader.EventChan():
			wantEvent := K('a')
			if event != wantEvent {
				t.Errorf("After Stop and Start, Reader reads %v, want %v", event, wantEvent)
			}
		case <-time.After(time.Second):
			t.Errorf("After Stop and Start, Reader timed out")
		}
	}
}

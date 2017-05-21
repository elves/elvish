package tty

import (
	"os"
	"testing"
	"time"

	"github.com/elves/elvish/edit/uitypes"
)

// timeout is the longest time the tests wait between writing something on
// the writer and reading it from the reader before declaring that the
// reader has a bug.
const timeoutInterval = 100 * time.Millisecond

func timeout() <-chan time.Time {
	return time.After(timeoutInterval)
}

var (
	writer *os.File
	reader *Reader
)

func TestMain(m *testing.M) {
	r, w, err := os.Pipe()
	if err != nil {
		panic("os.Pipe returned error, something is seriously wrong")
	}
	defer r.Close()
	defer w.Close()
	writer = w
	reader = NewReader(r)
	go reader.Run()
	defer reader.Quit()

	os.Exit(m.Run())
}

var keyTests = []struct {
	input string
	want  ReadUnit
}{
	// Simple graphical key.
	{"x", Key{'x', 0}},
	{"X", Key{'X', 0}},
	{" ", Key{' ', 0}},

	// Ctrl key.
	{"\001", Key{'A', uitypes.Ctrl}},
	{"\033", Key{'[', uitypes.Ctrl}},

	// Ctrl-ish keys, but not thought as Ctrl keys by our reader.
	{"\n", Key{'\n', 0}},
	{"\t", Key{'\t', 0}},
	{"\x7f", Key{'\x7f', 0}}, // backspace

	// Alt plus simple graphical key.
	{"\033a", Key{'a', uitypes.Alt}},
	{"\033[", Key{'[', uitypes.Alt}},

	// G3-style key.
	{"\033OA", Key{uitypes.Up, 0}},
	{"\033OH", Key{uitypes.Home, 0}},

	// CSI-sequence key identified by the ending rune.
	{"\033[A", Key{uitypes.Up, 0}},
	{"\033[H", Key{uitypes.Home, 0}},
	// Test for all possible modifier
	{"\033[1;2A", Key{uitypes.Up, uitypes.Shift}},

	// CSI-sequence key with one argument, always ending in '~'.
	{"\033[1~", Key{uitypes.Home, 0}},
	{"\033[11~", Key{uitypes.F1, 0}},

	// CSI-sequence key with three arguments and ending in '~'. The first
	// argument is always 27, the second identifies the modifier and the last
	// identifies the key.
	{"\033[27;4;63~", Key{';', uitypes.Shift | uitypes.Alt}},
}

func TestKey(t *testing.T) {
	for _, test := range keyTests {
		writer.WriteString(test.input)
		select {
		case k := <-reader.UnitChan():
			if k != test.want {
				t.Errorf("Reader reads key %v, want %v", k, test.want)
			}
		case <-timeout():
			t.Errorf("Reader fails to convert literal key")
		}
	}
}

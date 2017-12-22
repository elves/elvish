// +build !windows,!plan9

package edit

import (
	"testing"
	"time"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/sys"
	"github.com/kr/pty"
)

var readLineTests = []struct {
	input string
	want  string
}{
	{"\n", ""},
	{"test\n", "test"},
	{"abc\x7fd\n", "abd"},
	{"abc\x17d\n", "d"},
}

var readLineTimeout = 5 * time.Second

func TestReadLine(t *testing.T) {
	master, tty, err := pty.Open()
	if err != nil {
		panic(err)
	}
	defer master.Close()
	defer tty.Close()
	// Continually consume tty outputs so that the editor is not blocked on
	// writing.
	go func() {
		var buf [64]byte
		for {
			_, err := master.Read(buf[:])
			if err != nil {
				break
			}
		}
	}()

	ev := eval.NewEvaler()

	for _, test := range readLineTests {
		lineChan := make(chan string)
		errChan := make(chan error)
		go func() {
			ed := NewEditor(tty, tty, nil, ev)
			defer ed.Close()
			line, err := ed.ReadLine()
			if err != nil {
				errChan <- err
			} else {
				lineChan <- line
			}
		}()

		_, err := master.WriteString(test.input)
		if err != nil {
			panic(err)
		}

		select {
		case line := <-lineChan:
			if line != test.want {
				t.Errorf("ReadLine() => %q, want %q (input %q)", line, test.want, test.input)
			}
		case err := <-errChan:
			t.Errorf("ReadLine() => error %v (input %q)", err, test.input)
		case <-time.After(readLineTimeout):
			t.Errorf("ReadLine() timed out (input %q)", test.input)
			t.Log("\n" + sys.DumpStack())
		}
	}
}

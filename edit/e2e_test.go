// +build !windows,!plan9

// End-to-end tests for the editor. Only enabled on UNIX where pseudo-terminals
// are supported.

package edit

import (
	"io"
	"os"
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
	ev := eval.NewEvaler()
	defer ev.Close()

	for _, test := range readLineTests {
		// Editor output is only used in failure messages.
		var outputs []byte
		master, lineChan, errChan := run(ev, nil, &outputs)
		defer master.Close()

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
			t.Log("Stack trace: \n" + sys.DumpStack())
			t.Logf("Terminal output: %q", outputs)
			t.FailNow()
		}
	}
}

// run sets up a testing environment for an Editor, and calls its ReadLine
// method in a goroutine. It returns the master end of the pty the Editor is
// connected to, and two channels onto which the result of ReadLine will be
// delivered. The caller is responsible for closing the master file.
func run(ev *eval.Evaler, sigs <-chan os.Signal, ptrOutputs *[]byte) (*os.File,
	<-chan string, <-chan error) {

	master, tty, err := pty.Open()
	if err != nil {
		panic(err)
	}
	// Continually consume tty outputs so that the editor is not blocked on
	// writing.
	go drain(master, ptrOutputs)

	lineChan := make(chan string)
	errChan := make(chan error)

	go func() {
		ed := NewEditor(tty, tty, nil, ev)
		line, err := ed.ReadLine()
		if err != nil {
			errChan <- err
		} else {
			lineChan <- line
		}
		ed.Close()
		tty.Close()
		close(lineChan)
		close(errChan)
	}()

	return master, lineChan, errChan
}

// drain drains the given reader. If a non-nil []byte pointer is passed, it also
// makes the outputs available. It returns when r.Read returns an error.
func drain(r io.Reader, ptrOutputs *[]byte) {
	var buf [256]byte
	for {
		nr, err := r.Read(buf[:])
		if err != nil {
			return
		}
		if ptrOutputs != nil {
			*ptrOutputs = append(*ptrOutputs, buf[:nr]...)
		}
	}
}

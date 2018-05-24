// +build !windows,!plan9

// End-to-end tests for the editor. Only enabled on UNIX where pseudo-terminals
// are supported.

package edcore

import (
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/sys"
	"github.com/kr/pty"
)

// readLineTest contains the data for a test case of ReadLine.
type readLineTest struct {
	// The input to write to the master side.
	input []byte
	// Places where SIGINT should be sent to the editor, as indicies into the
	// input string. For example, if sigints is {1}, it means that a SIGINT
	// should be sent right after the first byte is sent.
	sigints []int
	// Expected line to be returned from ReadLine.
	wantLine string
}

func newTest(input string, wantLine string) *readLineTest {
	return &readLineTest{input: []byte(input), wantLine: wantLine}
}

func (t *readLineTest) sigint(x ...int) *readLineTest {
	t.sigints = x
	return t
}

var readLineTests = []*readLineTest{
	newTest("\n", ""),
	newTest("test\n", "test"),
	// \x7f is DEL and erases the previous character
	newTest("abc\x7fd\n", "abd"),
	// \x17 is ^U and erases the line before the cursor
	newTest("abc\x17d\n", "d"),
	// SIGINT resets the editor and erases the line. Disabled for now.
	// newTest("000123\n", "123").sigint(3),
}

var readLineTimeout = 5 * time.Second

func TestReadLine(t *testing.T) {
	ev := eval.NewEvaler()
	defer ev.Close()

	for _, test := range readLineTests {
		// Editor output is only used in failure messages.
		var outputs []byte
		sigs := make(chan os.Signal, 10)
		defer close(sigs)
		master, lineChan, errChan := run(ev, sigs, &outputs)
		defer master.Close()

		write(master, sigs, test.input, test.sigints)

		select {
		case line := <-lineChan:
			if line != test.wantLine {
				t.Errorf("ReadLine() => %q, want %q (input %q)", line, test.wantLine, test.input)
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

// write interprets the input and sigints arguments, and write inputs and
// signals to the writer and signal channel.
func write(w *os.File, sigs chan<- os.Signal, input []byte, sigints []int) {
	if len(sigints) == 0 {
		mustWrite(w, input)
		return
	}
	for i, idx := range sigints {
		lastidx := 0
		if i > 0 {
			lastidx = sigints[i-1]
		}
		mustWrite(w, input[lastidx:idx])
		sigs <- syscall.SIGINT
	}
	mustWrite(w, input[sigints[len(sigints)-1]:])
}

func mustWrite(w io.Writer, p []byte) {
	_, err := w.Write(p)
	if err != nil {
		panic(err)
	}
}

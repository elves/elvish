// Common testing utilities. This file does not file a _test.go suffix so that
// it can be used from other packages that also want to test the modules they
// implement (e.g. edit: and re:).

package eval

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
)

// Test is a test case for TestEval.
type Test struct {
	text string
	want
}

type want struct {
	out      []interface{}
	bytesOut []byte
	err      error
}

// A special value for want.err to indicate that any error, as long as not nil,
// is OK
var errAny = errors.New("any error")

// The following functions and methods are used to build Test structs. They are
// supposed to read like English, so a test that "put x" should put "x" reads:
//
// That("put x").Puts("x")

// That returns a new Test with the specified source code.
func That(text string) Test {
	return Test{text: text}
}

// DoesNothing returns t unchanged. It is used to mark that a piece of code
// should simply does nothing. In particular, it shouldn't have any output and
// does not error.
func (t Test) DoesNothing() Test {
	return t
}

// Puts returns an altered Test that requires the source code to produce the
// specified values in the value channel when evaluated.
func (t Test) Puts(vs ...interface{}) Test {
	t.want.out = vs
	return t
}

// Puts returns an altered Test that requires the source code to produce the
// specified strings in the value channel when evaluated.
func (t Test) PutsStrings(ss []string) Test {
	t.want.out = make([]interface{}, len(ss))
	for i, s := range ss {
		t.want.out[i] = s
	}
	return t
}

// Prints returns an altered test that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t Test) Prints(s string) Test {
	t.want.bytesOut = []byte(s)
	return t
}

// ErrorsWith returns an altered Test that requires the source code to result in
// the specified error when evaluted.
func (t Test) ErrorsWith(err error) Test {
	t.want.err = err
	return t
}

// Errors returns an altered Test that requires the source code to result in any
// error when evaluated.
func (t Test) Errors() Test {
	return t.ErrorsWith(errAny)
}

// RunTests runs test cases. For each test case, a new Evaler is made by calling
// makeEvaler.
func RunTests(t *testing.T, evalTests []Test, makeEvaler func() *Evaler) {
	for _, tt := range evalTests {
		// fmt.Printf("eval %q\n", tt.text)

		ev := makeEvaler()
		defer ev.Close()
		out, bytesOut, err := evalAndCollect(t, ev, []string{tt.text}, len(tt.want.out))

		first := true
		errorf := func(format string, args ...interface{}) {
			if first {
				first = false
				t.Errorf("eval(%q) fails:", tt.text)
			}
			t.Errorf("  "+format, args...)
		}

		if !matchOut(tt.want.out, out) {
			errorf("got out=%#v, want %#v", out, tt.want.out)
		}
		if !bytes.Equal(tt.want.bytesOut, bytesOut) {
			errorf("got bytesOut=%q, want %q", bytesOut, tt.want.bytesOut)
		}
		if !matchErr(tt.want.err, err) {
			errorf("got err=%v, want %v", err, tt.want.err)
		}
	}
}

func evalAndCollect(t *testing.T, ev *Evaler, texts []string, chsize int) ([]interface{}, []byte, error) {
	// Collect byte output
	bytesOut := []byte{}
	pr, pw, _ := os.Pipe()
	bytesDone := make(chan struct{})
	go func() {
		for {
			var buf [64]byte
			nr, err := pr.Read(buf[:])
			bytesOut = append(bytesOut, buf[:nr]...)
			if err != nil {
				break
			}
		}
		close(bytesDone)
	}()

	// Channel output
	outs := []interface{}{}

	// Eval error. Only that of the last text is saved.
	var ex error

	for i, text := range texts {
		name := fmt.Sprintf("test%d.elv", i)
		src := NewScriptSource(name, name, text)

		op := mustParseAndCompile(t, ev, src)

		outCh := make(chan interface{}, chsize)
		outDone := make(chan struct{})
		go func() {
			for v := range outCh {
				outs = append(outs, v)
			}
			close(outDone)
		}()

		ports := []*Port{
			{File: os.Stdin, Chan: ClosedChan},
			{File: pw, Chan: outCh},
			{File: os.Stderr, Chan: BlackholeChan},
		}

		ex = ev.eval(op, ports, src)
		close(outCh)
		<-outDone
	}

	pw.Close()
	<-bytesDone
	pr.Close()

	return outs, bytesOut, ex
}

func mustParseAndCompile(t *testing.T, ev *Evaler, src *Source) Op {
	n, err := parse.Parse(src.name, src.code)
	if err != nil {
		t.Fatalf("Parse(%q) error: %s", src.code, err)
	}
	op, err := ev.Compile(n, src)
	if err != nil {
		t.Fatalf("Compile(Parse(%q)) error: %s", src.code, err)
	}
	return op
}

func matchOut(want, got []interface{}) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !vals.Equal(got[i], want[i]) {
			return false
		}
	}
	return true
}

func matchErr(want, got error) bool {
	if got == nil {
		return want == nil
	}
	return want == errAny || reflect.DeepEqual(got.(*Exception).Cause, want)
}

// compareValues compares two slices, using equals for each element.
func compareSlice(wantValues, gotValues []interface{}) error {
	if len(wantValues) != len(gotValues) {
		return fmt.Errorf("want %d values, got %d",
			len(wantValues), len(gotValues))
	}
	for i, want := range wantValues {
		if !vals.Equal(want, gotValues[i]) {
			return fmt.Errorf("want [%d] = %s, got %s", i, want, gotValues[i])
		}
	}
	return nil
}

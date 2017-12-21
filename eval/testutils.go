// Common testing utilities. This file does not file a _test.go suffix so that
// it can be used from other packages that also want to test the modules they
// implement (e.g. edit: and re:).

package eval

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/parse"
)

// Test is a test case for TestEval.
type Test struct {
	text string
	want
}

type want struct {
	out      []Value
	bytesOut []byte
	err      error
}

// A special value for want.err to indicate that any error, as long as not nil,
// is OK
var errAny = errors.New("")

var (
	wantNothing = want{}
	wantTrue    = want{out: bools(true)}
	wantFalse   = want{out: bools(false)}
)

// Shorthands for values in want.out

func strs(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = String(s)
	}
	return vs
}

func bools(bs ...bool) []Value {
	vs := make([]Value, len(bs))
	for i, b := range bs {
		vs[i] = Bool(b)
	}
	return vs
}

// NewTest returns a new Test with the specified source code.
func NewTest(text string) Test {
	return Test{text: text}
}

// WantOut returns an altered Test that requires the source code to produce the
// specified values in the value channel when evaluated.
func (t Test) WantOut(vs ...Value) Test {
	t.want.out = vs
	return t
}

// WantOutStrings returns an altered Test that requires the source code to
// produce the specified string values in the value channel when evaluated.
func (t Test) WantOutStrings(ss ...string) Test {
	t.want.out = strs(ss...)
	return t
}

// WantOutBools returns an altered Test that requires the source code to produce
// the specified boolean values in the value channel when evaluated.
func (t Test) WantOutBools(bs ...bool) Test {
	t.want.out = bools(bs...)
	return t
}

// WantBytesOut returns an altered test that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t Test) WantBytesOut(b []byte) Test {
	t.want.bytesOut = b
	return t
}

// WantErr returns an altered Test that requires the source code to result in
// the specified error when evaluted.
func (t Test) WantErr(err error) Test {
	t.want.err = err
	return t
}

// WantAnyErr returns an altered Test that requires the source code to result in
// any error when evaluated.
func (t Test) WantAnyErr(err error) Test {
	t.want.err = err
	return t
}

func RunTests(t *testing.T, dataDir string, evalTests []Test) {
	for _, tt := range evalTests {
		// fmt.Printf("eval %q\n", tt.text)

		out, bytesOut, err := evalAndCollect(t, dataDir, []string{tt.text}, len(tt.want.out))

		first := true
		errorf := func(format string, args ...interface{}) {
			if first {
				first = false
				t.Errorf("eval(%q) fails:", tt.text)
			}
			t.Errorf("  "+format, args...)
		}

		if !matchOut(tt.want.out, out) {
			errorf("got out=%v, want %v", out, tt.want.out)
		}
		if string(tt.want.bytesOut) != string(bytesOut) {
			errorf("got bytesOut=%q, want %q", bytesOut, tt.want.bytesOut)
		}
		if !matchErr(tt.want.err, err) {
			errorf("got err=%v, want %v", err, tt.want.err)
		}
	}
}

func evalAndCollect(t *testing.T, dataDir string, texts []string, chsize int) ([]Value, []byte, error) {
	name := "<eval test>"
	ev := NewEvaler(api.NewClient("/invalid"), nil, dataDir, nil)

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
	outs := []Value{}

	// Eval error. Only that of the last text is saved.
	var ex error

	for _, text := range texts {
		op := mustParseAndCompile(t, ev, name, text)

		outCh := make(chan Value, chsize)
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

		ex = ev.eval(op, ports, name, text)
		close(outCh)
		<-outDone
	}

	pw.Close()
	<-bytesDone
	pr.Close()

	return outs, bytesOut, ex
}

func mustParseAndCompile(t *testing.T, ev *Evaler, name, text string) Op {
	n, err := parse.Parse(name, text)
	if err != nil {
		t.Fatalf("Parse(%q) error: %s", text, err)
	}
	op, err := ev.Compile(n, name, text)
	if err != nil {
		t.Fatalf("Compile(Parse(%q)) error: %s", text, err)
	}
	return op
}

func matchOut(want, got []Value) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
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
		if !equals(want, gotValues[i]) {
			return fmt.Errorf("want [%d] = %s, got %s", i, want, gotValues[i])
		}
	}
	return nil
}

// equals compares two values. It uses Eq if want is a Value instance, or
// reflect.DeepEqual otherwise.
func equals(a, b interface{}) bool {
	if aValue, ok := a.(Value); ok {
		return aValue.Equal(b)
	}
	return reflect.DeepEqual(a, b)
}

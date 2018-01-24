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

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
)

// Test is a test case for TestEval.
type Test struct {
	text string
	want
}

type want struct {
	out      []types.Value
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

func strs(ss ...string) []types.Value {
	vs := make([]types.Value, len(ss))
	for i, s := range ss {
		vs[i] = types.String(s)
	}
	return vs
}

func bools(bs ...bool) []types.Value {
	vs := make([]types.Value, len(bs))
	for i, b := range bs {
		vs[i] = types.Bool(b)
	}
	return vs
}

// NewTest returns a new Test with the specified source code.
func NewTest(text string) Test {
	return Test{text: text}
}

// WantOut returns an altered Test that requires the source code to produce the
// specified values in the value channel when evaluated.
func (t Test) WantOut(vs ...types.Value) Test {
	t.want.out = vs
	return t
}

// WantOutStrings returns an altered Test that requires the source code to
// produce the specified string values in the value channel when evaluated.
func (t Test) WantOutStrings(ss ...string) Test {
	return t.WantOut(strs(ss...)...)
}

// WantOutBools returns an altered Test that requires the source code to produce
// the specified boolean values in the value channel when evaluated.
func (t Test) WantOutBools(bs ...bool) Test {
	return t.WantOut(bools(bs...)...)
}

// WantBytesOut returns an altered test that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t Test) WantBytesOut(b []byte) Test {
	t.want.bytesOut = b
	return t
}

// WantBytesOutString is the same as WantBytesOut except that its argument is a
// string.
func (t Test) WantBytesOutString(s string) Test {
	return t.WantBytesOut([]byte(s))
}

// WantErr returns an altered Test that requires the source code to result in
// the specified error when evaluted.
func (t Test) WantErr(err error) Test {
	t.want.err = err
	return t
}

// WantAnyErr returns an altered Test that requires the source code to result in
// any error when evaluated.
func (t Test) WantAnyErr() Test {
	return t.WantErr(errAny)
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
			errorf("got out=%v, want %v", out, tt.want.out)
		}
		if !bytes.Equal(tt.want.bytesOut, bytesOut) {
			errorf("got bytesOut=%q, want %q", bytesOut, tt.want.bytesOut)
		}
		if !matchErr(tt.want.err, err) {
			errorf("got err=%v, want %v", err, tt.want.err)
		}
	}
}

func evalAndCollect(t *testing.T, ev *Evaler, texts []string, chsize int) ([]types.Value, []byte, error) {
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
	outs := []types.Value{}

	// Eval error. Only that of the last text is saved.
	var ex error

	for i, text := range texts {
		name := fmt.Sprintf("test%d.elv", i)
		src := NewScriptSource(name, name, text)

		op := mustParseAndCompile(t, ev, src)

		outCh := make(chan types.Value, chsize)
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

func matchOut(want, got []types.Value) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !types.Equal(got[i], want[i]) {
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
		if !types.Equal(want, gotValues[i]) {
			return fmt.Errorf("want [%d] = %s, got %s", i, want, gotValues[i])
		}
	}
	return nil
}

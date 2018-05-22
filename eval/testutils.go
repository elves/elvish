// Framework for testing Elvish script. This file does not file a _test.go
// suffix so that it can be used from other packages that also want to test the
// modules they implement (e.g. edit: and re:).
//
// The entry point for the framework is the Test function, which accepts a
// *testing.T and a variadic number of test cases. Test cases are constructed
// using the That function followed by methods that add constraints on the test
// case. Overall, a test looks like:
//
//     Test(t,
//         That("put x").Puts("x"),
//         That("echo x").Prints("x\n"))
//
// If some setup is needed, use the TestWithSetup function instead.

package eval

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// TestCase is a test case for Test.
type TestCase struct {
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
func That(text string) TestCase {
	return TestCase{text: text}
}

// DoesNothing returns t unchanged. It is used to mark that a piece of code
// should simply does nothing. In particular, it shouldn't have any output and
// does not error.
func (t TestCase) DoesNothing() TestCase {
	return t
}

// Puts returns an altered Test that requires the source code to produce the
// specified values in the value channel when evaluated.
func (t TestCase) Puts(vs ...interface{}) TestCase {
	t.want.out = vs
	return t
}

// Puts returns an altered Test that requires the source code to produce the
// specified strings in the value channel when evaluated.
func (t TestCase) PutsStrings(ss []string) TestCase {
	t.want.out = make([]interface{}, len(ss))
	for i, s := range ss {
		t.want.out[i] = s
	}
	return t
}

// Prints returns an altered test that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t TestCase) Prints(s string) TestCase {
	t.want.bytesOut = []byte(s)
	return t
}

// ErrorsWith returns an altered Test that requires the source code to result in
// the specified error when evaluted.
func (t TestCase) ErrorsWith(err error) TestCase {
	t.want.err = err
	return t
}

// Errors returns an altered Test that requires the source code to result in any
// error when evaluated.
func (t TestCase) Errors() TestCase {
	return t.ErrorsWith(errAny)
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler.
func Test(t *testing.T, tests ...TestCase) {
	TestWithSetup(t, func(*Evaler) {}, tests...)
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler and passed to the setup function.
func TestWithSetup(t *testing.T, setup func(*Evaler), tests ...TestCase) {
	for _, tt := range tests {
		ev := NewEvaler()
		setup(ev)
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

		ev.Close()
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

// MustMkdirAll calls os.MkdirAll and panics if an error is returned. It is
// mainly useful in tests.
func MustMkdirAll(name string, perm os.FileMode) {
	err := os.MkdirAll(name, perm)
	if err != nil {
		panic(err)
	}
}

// MustCreateEmpty creates an empty file, and panics if an error occurs. It is
// mainly useful in tests.
func MustCreateEmpty(name string) {
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	file.Close()
}

// MustWriteFile calls ioutil.WriteFile and panics if an error occurs. It is
// mainly useful in tests.
func MustWriteFile(filename string, data []byte, perm os.FileMode) {
	err := ioutil.WriteFile(filename, data, perm)
	if err != nil {
		panic(err)
	}
}

// InTempHome is like util.InTempDir, but it also sets HOME to the temporary
// directory when f is called.
func InTempHome(f func(string)) {
	util.InTempDir(func(tmpHome string) {
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		f(tmpHome)
		os.Setenv("HOME", oldHome)
	})
}

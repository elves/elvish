// Framework for testing Elvish script. This file does not have a _test.go
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
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
)

// TestCase is a test case for Test.
type TestCase struct {
	code string
	want result
}

type result struct {
	valueOut []interface{}
	bytesOut []byte

	compilationError error
	exception        error
}

// A special value for want.{compilationError,exception} to indicate that any
// error, as long as not nil, is a match.
var errAny = errors.New("any error")

// The following functions and methods are used to build Test structs. They are
// supposed to read like English, so a test that "put x" should put "x" reads:
//
// That("put x").Puts("x")

// That returns a new Test with the specified source code. Multiple arguments
// are joined with newlines.
func That(lines ...string) TestCase {
	return TestCase{code: strings.Join(lines, "\n")}
}

// DoesNothing returns t unchanged. It is used to mark that a piece of code
// should simply does nothing. In particular, it shouldn't have any output and
// does not error.
func (t TestCase) DoesNothing() TestCase {
	return t
}

// Puts returns an altered TestCase that requires the source code to produce the
// specified values in the value channel when evaluated.
func (t TestCase) Puts(vs ...interface{}) TestCase {
	t.want.valueOut = vs
	return t
}

// PutsStrings returns an altered TestCase that requires the source code to produce
// the specified strings in the value channel when evaluated.
func (t TestCase) PutsStrings(ss []string) TestCase {
	t.want.valueOut = make([]interface{}, len(ss))
	for i, s := range ss {
		t.want.valueOut[i] = s
	}
	return t
}

// Prints returns an altered TestCase that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t TestCase) Prints(s string) TestCase {
	t.want.bytesOut = []byte(s)
	return t
}

// Throws returns an altered TestCase that requires the source code to throw the
// specified exception when evaluted.
func (t TestCase) Throws(err error) TestCase {
	t.want.exception = err
	return t
}

// ThrowsAny returns an altered TestCase that requires the source code to throw
// any exception when evaluated.
func (t TestCase) ThrowsAny() TestCase {
	return t.Throws(errAny)
}

// DoesNotCompile returns an altered TestCase that requires the source code to
// fail compilation.
func (t TestCase) DoesNotCompile() TestCase {
	t.want.compilationError = errAny
	return t
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler.
func Test(t *testing.T, tests ...TestCase) {
	TestWithSetup(t, func(*Evaler) {}, tests...)
}

// TestWithSetup runs test cases. For each test case, a new Evaler is created
// with NewEvaler and passed to the setup function.
func TestWithSetup(t *testing.T, setup func(*Evaler), tests ...TestCase) {
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			ev := NewEvaler()
			defer ev.Close()
			setup(ev)

			r := evalAndCollect(t, ev, []string{tt.code})

			if !matchOut(tt.want.valueOut, r.valueOut) {
				t.Errorf("got value out %v, want %v", r.valueOut, tt.want.valueOut)
			}
			if !bytes.Equal(tt.want.bytesOut, r.bytesOut) {
				t.Errorf("got bytes out %q, want %q", r.bytesOut, tt.want.bytesOut)
			}
			if !matchErr(tt.want.compilationError, r.compilationError) {
				t.Errorf("got compilation error %v, want %v",
					r.compilationError, tt.want.compilationError)
			}
			if !matchErr(tt.want.exception, r.exception) {
				t.Errorf("got exception %v, want %v", r.exception, tt.want.exception)
			}
		})
	}
}

func evalAndCollect(t *testing.T, ev *Evaler, texts []string) result {
	var r result
	// Collect byte output.
	pr, pw, _ := os.Pipe()
	bytesDone := make(chan struct{})
	go func() {
		for {
			var buf [64]byte
			nr, err := pr.Read(buf[:])
			r.bytesOut = append(r.bytesOut, buf[:nr]...)
			if err != nil {
				break
			}
		}
		close(bytesDone)
	}()

	for i, text := range texts {
		name := fmt.Sprintf("test%d.elv", i)
		src := NewInternalElvishSource(true, name, text)

		n, err := parse.AsChunk(src.Name, src.Code)
		if err != nil {
			t.Fatalf("Parse(%q) error: %s", src.Code, err)
		}
		op, err := ev.Compile(n, src)
		if err != nil {
			// NOTE: Only the compilation error of the last code is saved.
			r.compilationError = err
			continue
		}

		outCh := make(chan interface{}, 1024)
		outDone := make(chan struct{})
		go func() {
			for v := range outCh {
				r.valueOut = append(r.valueOut, v)
			}
			close(outDone)
		}()

		ports := []*Port{
			{File: os.Stdin, Chan: ClosedChan},
			{File: pw, Chan: outCh},
			{File: os.Stderr, Chan: BlackholeChan},
		}

		// NOTE: Only the exception of the last code that compiles is saved.
		r.exception = ev.Eval(op, ports)
		close(outCh)
		<-outDone
	}

	pw.Close()
	<-bytesDone
	pr.Close()

	return r
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
	return want == errAny || reflect.DeepEqual(Cause(got), want)
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

// InTempHome is like util.InTestDir, but it also sets HOME to the temporary
// directory and restores the original HOME in cleanup.
//
// TODO(xiaq): Move this into the util package.
func InTempHome() (string, func()) {
	oldHome := os.Getenv("HOME")
	tmpHome, cleanup := util.InTestDir()
	os.Setenv("HOME", tmpHome)

	return tmpHome, func() {
		os.Setenv("HOME", oldHome)
		cleanup()
	}
}

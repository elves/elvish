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
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/elves/elvish/pkg/env"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
)

// These two symbols are used for tests that need to compare floating point
// values that can't be guaranteed to be bit for bit identical. Typically due
// to tiny rounding errors that tend to occur in floating point operations.
const float64EqualityThreshold = 1e-15

type Approximately struct{ F float64 }

// TestCase is a test case for Test.
type TestCase struct {
	code string
	want result
}

type result struct {
	valueOut   []interface{}
	bytesOut   []byte
	stderrPart []byte

	compilationError error
	exception        error
}

type errorMatcher interface{ matchError(error) bool }

// An errorMatcher for any error.
type anyError struct{}

func (anyError) Error() string { return "any error" }

func (anyError) matchError(e error) bool { return e != nil }

// An errorMatcher for any exception with the given cause and stack traces.
type exc struct {
	cause  error
	stacks []string
}

func (e exc) Error() string {
	return fmt.Sprintf("exception with cause %v and stacks %v", e.cause, e.stacks)
}

func (e exc) matchError(e2 error) bool {
	if e2, ok := e2.(*Exception); ok {
		if matchErr(e.cause, e2.Reason) {
			return reflect.DeepEqual(e.stacks, getStackTexts(e2.StackTrace))
		}
	}
	return false
}

func getStackTexts(tb *stackTrace) []string {
	texts := []string{}
	for tb != nil {
		ctx := tb.head
		texts = append(texts, ctx.Source[ctx.From:ctx.To])
		tb = tb.next
	}
	return texts
}

// An errorMatcher for any exception with the given cause.
type excWithCause struct{ cause error }

func (e excWithCause) Error() string { return "exception with cause " + e.cause.Error() }

func (e excWithCause) matchError(e2 error) bool {
	return e2 != nil && reflect.DeepEqual(e.cause, Cause(e2))
}

// An errorMatcher for any error with the given type.
type errWithType struct{ v error }

func (e errWithType) Error() string { return fmt.Sprintf("error with type %T", e.v) }

func (e errWithType) matchError(e2 error) bool {
	return reflect.TypeOf(e.v) == reflect.TypeOf(e2)
}

// An errorMatcher for any error with the given message.
type errWithMessage struct{ msg string }

func (e errWithMessage) Error() string { return "error with message " + e.msg }

func (e errWithMessage) matchError(e2 error) bool {
	return e2 != nil && e.msg == e2.Error()
}

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

// Prints returns an altered TestCase that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t TestCase) Prints(s string) TestCase {
	t.want.bytesOut = []byte(s)
	return t
}

// PrintsStderr returns an altered TestCase that requires the stderr output to
// contain the given text.
func (t TestCase) PrintsStderrWith(s string) TestCase {
	t.want.stderrPart = []byte(s)
	return t
}

// Throws returns an altered TestCase that requires the source code to throw an
// exception that has the given cause, and has stacktraces that match the given
// source fragments (innermost first).
func (t TestCase) Throws(cause error, stacks ...string) TestCase {
	return t.throws(exc{cause, stacks})
}

// ThrowsCause returns an altered TestCase that requires the source code to
// throw an exception with the given cause when evaluated.
//
// If the stack trace is important, use Throws instead of this method.
func (t TestCase) ThrowsCause(err error) TestCase {
	return t.throws(excWithCause{err})
}

// ThrowsMessage returns an altered TestCase that requires the source code to
// throw an error with the specified message when evaluted.
//
// Whenever possible, use ThrowsCause instead of this method.
func (t TestCase) ThrowsMessage(msg string) TestCase {
	return t.throws(errWithMessage{msg})
}

// ThrowsAny returns an altered TestCase that requires the source code to throw
// any exception when evaluated.
//
// Whenever possible, use a more specific Throws* method instead of this method.
func (t TestCase) ThrowsAny() TestCase {
	return t.throws(anyError{})
}

func (t TestCase) throws(err error) TestCase {
	t.want.exception = err
	return t
}

// DoesNotCompile returns an altered TestCase that requires the source code to
// fail compilation.
func (t TestCase) DoesNotCompile() TestCase {
	t.want.compilationError = anyError{}
	return t
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler.
func Test(t *testing.T, tests ...TestCase) {
	t.Helper()
	TestWithSetup(t, func(*Evaler) {}, tests...)
}

// TestWithSetup runs test cases. For each test case, a new Evaler is created
// with NewEvaler and passed to the setup function.
func TestWithSetup(t *testing.T, setup func(*Evaler), tests ...TestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Helper()
			ev := NewEvaler()
			defer ev.Close()
			setup(ev)

			r := evalAndCollect(t, ev, []string{tt.code})

			if !matchOut(tt.want.valueOut, r.valueOut) {
				t.Errorf("got value out %v, want %v",
					reprs(r.valueOut), reprs(tt.want.valueOut))
			}
			if !bytes.Equal(tt.want.bytesOut, r.bytesOut) {
				t.Errorf("got bytes out %q, want %q", r.bytesOut, tt.want.bytesOut)
			}
			if !bytes.Contains(r.stderrPart, tt.want.stderrPart) {
				t.Errorf("got stderr out %q, want %q", r.stderrPart, tt.want.stderrPart)
			}
			if !matchErr(tt.want.compilationError, r.compilationError) {
				t.Errorf("got compilation error %v, want %v",
					r.compilationError, tt.want.compilationError)
			}
			if !matchErr(tt.want.exception, r.exception) {
				t.Errorf("unexpected exception")
				t.Logf("got: %v", r.exception)
				if exc, ok := r.exception.(*Exception); ok {
					t.Logf("stack trace: %#v", getStackTexts(exc.StackTrace))
				}
				t.Errorf("want: %v", tt.want.exception)
			}
		})
	}
}

func evalAndCollect(t *testing.T, ev *Evaler, texts []string) result {
	var r result

	var wg sync.WaitGroup
	wg.Add(3)
	rOut, stdout := mustPipe()
	go func() {
		r.bytesOut = mustReadAllAndClose(rOut)
		wg.Done()
	}()
	rErr, stderr := mustPipe()
	go func() {
		r.stderrPart = mustReadAllAndClose(rErr)
		wg.Done()
	}()
	outCh := make(chan interface{}, 1024)
	go func() {
		for v := range outCh {
			r.valueOut = append(r.valueOut, v)
		}
		wg.Done()
	}()
	ports := []*Port{
		DevNullClosedChan,
		{File: stdout, Chan: outCh},
		{File: stderr, Chan: BlackholeChan},
	}

	for _, text := range texts {
		src := parse.Source{Name: "[test]", Code: text}

		tree, err := parse.Parse(src)
		if err != nil {
			t.Fatalf("Parse(%q) error: %s", src.Code, err)
		}
		op, err := ev.Compile(tree, nil)
		if err != nil {
			// NOTE: Only the compilation error of the last code is saved.
			r.compilationError = err
			continue
		}
		// NOTE: Only the exception of the last code that compiles is saved.
		r.exception = ev.Eval(op, EvalCfg{Ports: ports, Interrupt: ListenInterrupts})
	}

	stdout.Close()
	stderr.Close()
	close(outCh)
	wg.Wait()

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
		// Equality of some data types needs to be special-cased in unit
		// tests. For example, by definition `NaN == NaN` is always false
		// since NaN is never equal to any other value; not even NaN. But for
		// unit tests we want to ensure that if the test is expected to
		// produce NaN it does so and the test passes.
		switch v := got[i].(type) {
		case float64:
			switch x := want[i].(type) {
			case float64:
				if math.IsNaN(v) && math.IsNaN(x) {
					return true
				}
				return v == x
			case Approximately:
				// Apply a reasonable epsilon if the user asked for an
				// approximate equality test.
				w := x.F
				if math.IsNaN(v) && math.IsNaN(w) {
					return true
				}
				if math.IsInf(v, 0) && math.IsInf(w, 0) &&
					math.Signbit(v) == math.Signbit(w) {
					return true
				}
				return math.Abs(v-w) <= float64EqualityThreshold
			}
		}

		if !vals.Equal(got[i], want[i]) {
			return false
		}
	}
	return true
}

func reprs(values []interface{}) []string {
	s := make([]string, len(values))
	for i, v := range values {
		s[i] = vals.Repr(v, vals.NoPretty)
	}
	return s
}

func matchErr(want, got error) bool {
	if want == nil {
		return got == nil
	}
	if matcher, ok := want.(errorMatcher); ok {
		return matcher.matchError(got)
	}
	return reflect.DeepEqual(want, got)
}

func mustPipe() (*os.File, *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w
}

func mustReadAllAndClose(r io.ReadCloser) []byte {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	r.Close()
	return bs
}

// Calls os.MkdirAll and panics if an error is returned.
func mustMkdirAll(names ...string) {
	for _, name := range names {
		err := os.MkdirAll(name, 0700)
		if err != nil {
			panic(err)
		}
	}
}

// Creates an empty file, and panics if an error occurs.
func mustCreateEmpty(names ...string) {
	for _, name := range names {
		file, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

// Calls ioutil.WriteFile and panics if an error occurs.
func mustWriteFile(filename string, data []byte, perm os.FileMode) {
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
	oldHome := os.Getenv(env.HOME)
	tmpHome, cleanup := util.InTestDir()
	os.Setenv(env.HOME, tmpHome)

	return tmpHome, func() {
		os.Setenv(env.HOME, oldHome)
		cleanup()
	}
}

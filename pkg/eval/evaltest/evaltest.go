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

package evaltest

import (
	"bytes"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/testutil"
)

// TestCase is a test case for Test.
type TestCase struct {
	code string
	want Result
}

type Result struct {
	ValueOut  []interface{}
	BytesOut  []byte
	StderrOut []byte

	CompilationError error
	Exception        error
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
	t.want.ValueOut = vs
	return t
}

// Prints returns an altered TestCase that requires the source code to produce
// the specified output in the byte pipe when evaluated.
func (t TestCase) Prints(s string) TestCase {
	t.want.BytesOut = []byte(s)
	return t
}

// PrintsStderrWith returns an altered TestCase that requires the stderr
// output to contain the given text.
func (t TestCase) PrintsStderrWith(s string) TestCase {
	t.want.StderrOut = []byte(s)
	return t
}

// Throws returns an altered TestCase that requires the source code to throw an
// exception with the given reason. The reason supports special matcher values
// constructed by functions like ErrorWithMessage.
//
// If at least one stacktrace string is given, the exception must also have a
// stacktrace matching the given source fragments, frame by frame (innermost
// frame first). If no stacktrace string is given, the stack trace of the
// exception is not checked.
func (t TestCase) Throws(reason error, stacks ...string) TestCase {
	t.want.Exception = exc{reason, stacks}
	return t
}

// DoesNotCompile returns an altered TestCase that requires the source code to
// fail compilation.
func (t TestCase) DoesNotCompile() TestCase {
	t.want.CompilationError = anyError{}
	return t
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler.
func Test(t *testing.T, tests ...TestCase) {
	t.Helper()
	TestWithSetup(t, func(*eval.Evaler) {}, tests...)
}

// TestWithSetup runs test cases. For each test case, a new Evaler is created
// with NewEvaler and passed to the setup function.
func TestWithSetup(t *testing.T, setup func(*eval.Evaler), tests ...TestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Helper()
			ev := eval.NewEvaler()
			defer ev.Close()
			setup(ev)

			r := EvalAndCollect(t, ev, []string{tt.code})

			if !matchOut(tt.want.ValueOut, r.ValueOut) {
				t.Errorf("got value out %v, want %v",
					reprs(r.ValueOut), reprs(tt.want.ValueOut))
			}
			if !bytes.Equal(tt.want.BytesOut, r.BytesOut) {
				t.Errorf("got bytes out %q, want %q", r.BytesOut, tt.want.BytesOut)
			}
			if !bytes.Contains(r.StderrOut, tt.want.StderrOut) {
				t.Errorf("got stderr out %q, want %q", r.StderrOut, tt.want.StderrOut)
			}
			if !matchErr(tt.want.CompilationError, r.CompilationError) {
				t.Errorf("got compilation error %v, want %v",
					r.CompilationError, tt.want.CompilationError)
			}
			if !matchErr(tt.want.Exception, r.Exception) {
				t.Errorf("unexpected exception")
				t.Logf("got: %v", r.Exception)
				if exc, ok := r.Exception.(*eval.Exception); ok {
					t.Logf("stack trace: %#v", getStackTexts(exc.StackTrace))
				}
				t.Errorf("want: %v", tt.want.Exception)
			}
		})
	}
}

func EvalAndCollect(t *testing.T, ev *eval.Evaler, texts []string) Result {
	var r Result

	var wg sync.WaitGroup
	wg.Add(3)
	rOut, stdout := testutil.MustPipe()
	go func() {
		r.BytesOut = testutil.MustReadAllAndClose(rOut)
		wg.Done()
	}()
	rErr, stderr := testutil.MustPipe()
	go func() {
		r.StderrOut = testutil.MustReadAllAndClose(rErr)
		wg.Done()
	}()
	outCh := make(chan interface{}, 1024)
	go func() {
		for v := range outCh {
			r.ValueOut = append(r.ValueOut, v)
		}
		wg.Done()
	}()
	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{File: stdout, Chan: outCh},
		{File: stderr, Chan: eval.BlackholeChan},
	}

	for _, text := range texts {
		err := ev.Eval(parse.Source{Name: "[test]", Code: text},
			eval.EvalCfg{Ports: ports, Interrupt: eval.ListenInterrupts})

		if parse.GetError(err) != nil {
			t.Fatalf("Parse(%q) error: %s", text, err)
		} else if eval.GetCompilationError(err) != nil {
			// NOTE: Only the compilation error of the last code is saved.
			r.CompilationError = err
		} else if err != nil {
			// NOTE: Only the exception of the last code that compiles is saved.
			r.Exception = err
		}
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
		switch g := got[i].(type) {
		case float64:
			// Special-case float64 to correctly handle NaN and support
			// approximate comparison.
			switch w := want[i].(type) {
			case float64:
				if !matchFloat64(g, w, 0) {
					return false
				}
			case Approximately:
				if !matchFloat64(g, w.F, ApproximatelyThreshold) {
					return false
				}
			default:
				return false
			}
		default:
			if !vals.Equal(got[i], want[i]) {
				return false
			}
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

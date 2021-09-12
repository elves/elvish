// Package evaltest provides a framework for testing Elvish script.
//
// The entry point for the framework is the Test function, which accepts a
// *testing.T and any number of test cases.
//
// Test cases are constructed using the That function, followed by method calls
// that add additional information to it.
//
// Example:
//
//     Test(t,
//         That("put x").Puts("x"),
//         That("echo x").Prints("x\n"))
//
// If some setup is needed, use the TestWithSetup function instead.

package evaltest

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

// TestCase is a test case for Test.
type TestCase struct {
	codes []string
	want  Result
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
// are joined with newlines. To specify multiple pieces of code that are
// executed separately, use the Then method to append code pieces.
func That(lines ...string) TestCase {
	return TestCase{codes: []string{strings.Join(lines, "\n")}}
}

// Then returns a new Test that executes the given code in addition. Multiple
// arguments are joined with newlines.
func (t TestCase) Then(lines ...string) TestCase {
	t.codes = append(t.codes, strings.Join(lines, "\n"))
	return t
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
		t.Run(strings.Join(tt.codes, "\n"), func(t *testing.T) {
			t.Helper()
			ev := eval.NewEvaler()
			setup(ev)

			r := evalAndCollect(t, ev, tt.codes)

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
				if exc, ok := r.Exception.(eval.Exception); ok {
					// For an eval.Exception report the type of the underlying error.
					t.Logf("got: %T: %v", exc.Reason(), exc)
					t.Logf("stack trace: %#v", getStackTexts(exc.StackTrace()))
				} else {
					t.Logf("got: %T: %v", r.Exception, r.Exception)
				}
				t.Errorf("want: %v", tt.want.Exception)
			}
		})
	}
}

func evalAndCollect(t *testing.T, ev *eval.Evaler, texts []string) Result {
	var r Result

	port1, collect1 := capturePort()
	port2, collect2 := capturePort()
	ports := []*eval.Port{eval.DummyInputPort, port1, port2}

	for _, text := range texts {
		err := ev.Eval(parse.Source{Name: "[test]", Code: text},
			eval.EvalCfg{Ports: ports, Interrupt: eval.ListenInterrupts})

		if parse.GetError(err) != nil {
			t.Fatalf("Parse(%q) error: %s", text, err)
		} else if eval.GetCompilationError(err) != nil {
			// NOTE: If multiple code pieces have compilation errors, only the
			// last one compilation error is saved.
			r.CompilationError = err
		} else if err != nil {
			// NOTE: If multiple code pieces throw exceptions, only the last one
			// is saved.
			r.Exception = err
		}
	}

	r.ValueOut, r.BytesOut = collect1()
	_, r.StderrOut = collect2()
	return r
}

// Like eval.CapturePort, but captures values and bytes separately. Also panics
// if it cannot create a pipe.
func capturePort() (*eval.Port, func() ([]interface{}, []byte)) {
	var values []interface{}
	var bytes []byte
	port, done, err := eval.PipePort(
		func(ch <-chan interface{}) {
			for v := range ch {
				values = append(values, v)
			}
		},
		func(r *os.File) {
			bytes = testutil.MustReadAllAndClose(r)
		})
	if err != nil {
		panic(err)
	}
	return port, func() ([]interface{}, []byte) {
		done()
		return values, bytes
	}
}

func matchOut(want, got []interface{}) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !match(got[i], want[i]) {
			return false
		}
	}
	return true
}

func match(got, want interface{}) bool {
	switch got := got.(type) {
	case float64:
		// Special-case float64 to correctly handle NaN and support
		// approximate comparison.
		switch want := want.(type) {
		case float64:
			return matchFloat64(got, want, 0)
		case Approximately:
			return matchFloat64(got, want.F, ApproximatelyThreshold)
		}
	case string:
		switch want := want.(type) {
		case MatchingRegexp:
			return matchRegexp(want.Pattern, got)
		}
	}
	return vals.Equal(got, want)
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

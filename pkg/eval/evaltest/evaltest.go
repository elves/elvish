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

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

// Case is a test case that can be used in Test.
type Case struct {
	codes  []string
	setup  func(ev *eval.Evaler)
	verify func(t *testing.T)
	want   result
}

type result struct {
	ValueOut  []any
	BytesOut  []byte
	StderrOut []byte

	CompilationError error
	Exception        error
}

// That returns a new Case with the specified source code. Multiple arguments
// are joined with newlines. To specify multiple pieces of code that are
// executed separately, use the Then method to append code pieces.
//
// When combined with subsequent method calls, a test case reads like English.
// For example, a test for the fact that "put x" puts "x" reads:
//
//     That("put x").Puts("x")
func That(lines ...string) Case {
	return Case{codes: []string{strings.Join(lines, "\n")}}
}

// Then returns a new Case that executes the given code in addition. Multiple
// arguments are joined with newlines.
func (c Case) Then(lines ...string) Case {
	c.codes = append(c.codes, strings.Join(lines, "\n"))
	return c
}

// Then returns a new Case with the given setup function executed on the Evaler
// before the code is executed.
func (c Case) WithSetup(f func(*eval.Evaler)) Case {
	c.setup = f
	return c
}

// DoesNothing returns t unchanged. It is useful to mark tests that don't have
// any side effects, for example:
//
//     That("nop").DoesNothing()
func (c Case) DoesNothing() Case {
	return c
}

// Puts returns an altered Case that runs an additional verification function.
func (c Case) Passes(f func(t *testing.T)) Case {
	c.verify = f
	return c
}

// Puts returns an altered Case that requires the source code to produce the
// specified values in the value channel when evaluated.
func (c Case) Puts(vs ...any) Case {
	c.want.ValueOut = vs
	return c
}

// Prints returns an altered Case that requires the source code to produce the
// specified output in the byte pipe when evaluated.
func (c Case) Prints(s string) Case {
	c.want.BytesOut = []byte(s)
	return c
}

// PrintsStderrWith returns an altered Case that requires the stderr output to
// contain the given text.
func (c Case) PrintsStderrWith(s string) Case {
	c.want.StderrOut = []byte(s)
	return c
}

// Throws returns an altered Case that requires the source code to throw an
// exception with the given reason. The reason supports special matcher values
// constructed by functions like ErrorWithMessage.
//
// If at least one stacktrace string is given, the exception must also have a
// stacktrace matching the given source fragments, frame by frame (innermost
// frame first). If no stacktrace string is given, the stack trace of the
// exception is not checked.
func (c Case) Throws(reason error, stacks ...string) Case {
	c.want.Exception = exc{reason, stacks}
	return c
}

// DoesNotCompile returns an altered Case that requires the source code to fail
// compilation.
func (c Case) DoesNotCompile() Case {
	c.want.CompilationError = anyError{}
	return c
}

// Test runs test cases. For each test case, a new Evaler is created with
// NewEvaler.
func Test(t *testing.T, tests ...Case) {
	t.Helper()
	TestWithSetup(t, func(*eval.Evaler) {}, tests...)
}

// TestWithSetup runs test cases. For each test case, a new Evaler is created
// with NewEvaler and passed to the setup function.
func TestWithSetup(t *testing.T, setup func(*eval.Evaler), tests ...Case) {
	t.Helper()
	for _, tc := range tests {
		t.Run(strings.Join(tc.codes, "\n"), func(t *testing.T) {
			t.Helper()
			ev := eval.NewEvaler()
			setup(ev)
			if tc.setup != nil {
				tc.setup(ev)
			}

			r := evalAndCollect(t, ev, tc.codes)

			if tc.verify != nil {
				tc.verify(t)
			}
			if !matchOut(tc.want.ValueOut, r.ValueOut) {
				t.Errorf("got value out (-want +got):\n%s",
					cmp.Diff(r.ValueOut, tc.want.ValueOut, tt.CommonCmpOpt))
			}
			if !bytes.Equal(tc.want.BytesOut, r.BytesOut) {
				t.Errorf("got bytes out %q, want %q", r.BytesOut, tc.want.BytesOut)
			}
			if tc.want.StderrOut == nil {
				if len(r.StderrOut) > 0 {
					t.Errorf("got stderr out %q, want empty", r.StderrOut)
				}
			} else {
				if !bytes.Contains(r.StderrOut, tc.want.StderrOut) {
					t.Errorf("got stderr out %q, want output containing %q",
						r.StderrOut, tc.want.StderrOut)
				}
			}
			if !matchErr(tc.want.CompilationError, r.CompilationError) {
				t.Errorf("got compilation error %v, want %v",
					r.CompilationError, tc.want.CompilationError)
			}
			if !matchErr(tc.want.Exception, r.Exception) {
				t.Errorf("unexpected exception")
				if exc, ok := r.Exception.(eval.Exception); ok {
					// For an eval.Exception report the type of the underlying error.
					t.Logf("got: %T: %v", exc.Reason(), exc)
					t.Logf("stack trace: %#v", getStackTexts(exc.StackTrace()))
				} else {
					t.Logf("got: %T: %v", r.Exception, r.Exception)
				}
				t.Errorf("want: %v", tc.want.Exception)
			}
		})
	}
}

func evalAndCollect(t *testing.T, ev *eval.Evaler, texts []string) result {
	var r result

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
func capturePort() (*eval.Port, func() ([]any, []byte)) {
	var values []any
	var bytes []byte
	port, done, err := eval.PipePort(
		func(ch <-chan any) {
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
	return port, func() ([]any, []byte) {
		done()
		return values, bytes
	}
}

func matchOut(want, got []any) bool {
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

func match(got, want any) bool {
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

func matchErr(want, got error) bool {
	if want == nil {
		return got == nil
	}
	if matcher, ok := want.(errorMatcher); ok {
		return matcher.matchError(got)
	}
	return reflect.DeepEqual(want, got)
}

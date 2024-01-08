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
//	Test(t,
//	    That("put x").Puts("x"),
//	    That("echo x").Prints("x\n"))
//
// If some setup is needed, use the TestWithSetup function instead.
package evaltest

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/parse"
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
//	That("put x").Puts("x")
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
//	That("nop").DoesNothing()
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
// compilation with the given error messages.
func (c Case) DoesNotCompile(msgs ...string) Case {
	c.want.CompilationError = compilationError{msgs}
	return c
}

// Test is a shorthand for [TestWithSetup] when no setup is needed.
func Test(t *testing.T, tests ...Case) {
	t.Helper()
	TestWithSetup(t, func(*testing.T, *eval.Evaler) {}, tests...)
}

// TestWithEvalerSetup is a shorthand for [TestWithSetup] when the setup only
// needs to manipulate [eval.Evaler].
func TestWithEvalerSetup(t *testing.T, setup func(*eval.Evaler), tests ...Case) {
	t.Helper()
	TestWithSetup(t, func(_ *testing.T, ev *eval.Evaler) { setup(ev) }, tests...)
}

// TestWithSetup runs test cases.
//
// Each test case is run as a subtest with a newly created Evaler. The setup
// function is called with the [testing.T] and the [eval.Evaler] for the subset
// before code evaluation.
func TestWithSetup(t *testing.T, setup func(*testing.T, *eval.Evaler), tests ...Case) {
	t.Helper()
	for _, tc := range tests {
		t.Run(strings.Join(tc.codes, "\n"), func(t *testing.T) {
			t.Helper()
			ev := eval.NewEvaler()
			setup(t, ev)
			if tc.setup != nil {
				tc.setup(ev)
			}

			r := evalAndCollect(t, ev, tc.codes)

			if tc.verify != nil {
				tc.verify(t)
			}
			if !matchOut(tc.want.ValueOut, r.ValueOut) {
				t.Errorf("got value out (-want +got):\n%s",
					cmp.Diff(reprValues(tc.want.ValueOut), reprValues(r.ValueOut)))
			}
			if !bytes.Equal(tc.want.BytesOut, r.BytesOut) {
				t.Errorf("got bytes out (-want +got):\n%s",
					cmp.Diff(string(tc.want.BytesOut), string(r.BytesOut)))
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
				t.Errorf("got compilation error:\n%v\nwant %v",
					show(r.CompilationError), tc.want.CompilationError)
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

func reprValues(xs []any) []string {
	rs := make([]string, len(xs))
	for i, x := range xs {
		rs[i] = vals.Repr(x, 0)
	}
	return rs
}

// Use returns a function simulates "use" on an Evaler. Arguments must come in
// (string, eval.Nser) pairs.
func Use(args ...any) func(*eval.Evaler) {
	if len(args)%2 != 0 {
		panic("odd number of arguments")
	}
	ns := eval.BuildNs()
	for i := 0; i < len(args); i += 2 {
		ns.AddNs(args[i].(string), args[i+1].(eval.Nser))
	}
	return func(ev *eval.Evaler) {
		ev.ExtendGlobal(ns)
	}
}

func evalAndCollect(t *testing.T, ev *eval.Evaler, texts []string) result {
	var r result

	port1, collect1 := must.OK2(eval.CapturePort())
	port2, collect2 := must.OK2(eval.CapturePort())
	ports := []*eval.Port{eval.DummyInputPort, port1, port2}

	for _, text := range texts {
		ctx, done := eval.ListenInterrupts()
		err := ev.Eval(parse.Source{Name: "[test]", Code: text},
			eval.EvalCfg{Ports: ports, Interrupts: ctx})
		done()

		if parse.UnpackErrors(err) != nil {
			t.Fatalf("Parse(%q) error: %s", text, err)
		} else if eval.UnpackCompilationErrors(err) != nil {
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
	if matcher, ok := want.(ValueMatcher); ok {
		return matcher.matchValue(got)
	}
	// Special-case float64 to handle NaNs and infinities.
	if got, ok := got.(float64); ok {
		if want, ok := want.(float64); ok {
			return matchFloat64(got, want, 0)
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

func show(v any) string {
	if s, ok := v.(diag.Shower); ok {
		return s.Show("")
	}
	return fmt.Sprint(v)
}

// Package tt supports table-driven tests with little boilerplate.
//
// A typical use of this package looks like this:
//
//	// Function being tested
//	func Neg(i int) { return -i }
//
//	func TestNeg(t *testing.T) {
//		Test(t, Neg,
//			// Unnamed test case
//			Args(1).Rets(-1),
//			// Named test case
//			It("returns 0 for 0").Args(0).Rets(0),
//		)
//	}
package tt

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Case represents a test case. It has setter methods that augment and return
// itself, so they can be chained like It(...).Args(...).Rets(...).
type Case struct {
	fileAndLine  string
	desc         string
	args         []any
	retsMatchers [][]any
}

// It returns a Case with the given text description.
func It(desc string) *Case {
	return &Case{fileAndLine: fileAndLine(2), desc: desc}
}

// Args is equivalent to It("").args(...). It is useful when the test case is
// trivial and doesn't need a description; for more complex or interesting test
// cases, use [It] instead.
func Args(args ...any) *Case {
	return &Case{fileAndLine: fileAndLine(2), args: args}
}

func fileAndLine(skip int) string {
	_, filename, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s:%d", filepath.Base(filename), line)
}

// Args modifies the Case to pass the given arguments. It returns the receiver.
func (c *Case) Args(args ...any) *Case {
	c.args = args
	return c
}

// Rets modifies the Case to expect the given return values. It returns the
// receiver.
//
// The arguments may implement the [Matcher] interface, in which case its Match
// method is called with the actual return value. Otherwise, [reflect.DeepEqual]
// is used to determine matches.
func (c *Case) Rets(matchers ...any) *Case {
	c.retsMatchers = append(c.retsMatchers, matchers)
	return c
}

// FnDescriptor describes a function to test. It has setter methods that augment
// and return itself, so they can be chained like
// Fn(...).Named(...).ArgsFmt(...).
type FnDescriptor struct {
	name    string
	body    any
	argsFmt string
}

// Fn creates a FnDescriptor for the given function.
func Fn(body any) *FnDescriptor {
	return &FnDescriptor{body: body}
}

// Named sets the name of the function. This is only necessary for methods and
// local closures; package-level functions will have their name automatically
// inferred via reflection. It returns the receiver.
func (fn *FnDescriptor) Named(name string) *FnDescriptor {
	fn.name = name
	return fn
}

// ArgsFmt sets the string for formatting arguments in test error messages. It
// returns the receiver.
func (fn *FnDescriptor) ArgsFmt(s string) *FnDescriptor {
	fn.argsFmt = s
	return fn
}

// Test tests fn against the given Case instances.
//
// The fn argument may be the function itself or an explicit [FnDescriptor], the
// former case being equivalent to passing Fn(fn).
func Test(t *testing.T, fn any, tests ...*Case) {
	testInner[*testing.T](t, fn, tests...)
}

// Instead of using [*testing.T] directly, the inner implementation uses two
// interfaces so that it can be mocked. We need two interfaces because
// type parameters can't refer to the type itself.

type testRunner[T subtestRunner] interface {
	Helper()
	Run(name string, f func(t T)) bool
}

type subtestRunner interface {
	Errorf(format string, args ...any)
}

func testInner[T subtestRunner](t testRunner[T], fn any, tests ...*Case) {
	t.Helper()
	var fnd *FnDescriptor
	switch fn := fn.(type) {
	case *FnDescriptor:
		fnd = &FnDescriptor{}
		*fnd = *fn
	default:
		fnd = Fn(fn)
	}
	if fnd.name == "" {
		// Use reflection to discover the function's name.
		name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
		// Tests are usually restricted to functions in the same package, so
		// elide the package name.
		if i := strings.LastIndexByte(name, '.'); i != -1 {
			name = name[i+1:]
		}
		fnd.name = name
	}

	for _, test := range tests {
		t.Run(test.desc, func(t T) {
			rets := call(fnd.body, test.args)
			for _, retsMatcher := range test.retsMatchers {
				if !match(retsMatcher, rets) {
					var args string
					if fnd.argsFmt == "" {
						args = sprintArgs(test.args...)
					} else {
						args = fmt.Sprintf(fnd.argsFmt, test.args...)
					}
					var diff string
					if len(retsMatcher) == 1 && len(rets) == 1 {
						diff = cmp.Diff(retsMatcher[0], rets[0], cmpopt)
					} else {
						diff = cmp.Diff(retsMatcher, rets, cmpopt)
					}
					t.Errorf("%s: %s(%s) returns (-want +got):\n%s",
						test.fileAndLine, fnd.name, args, diff)
				}
			}
		})
	}
}

// RetValue is an empty interface used in the [Matcher] interface.
type RetValue any

// Matcher wraps the Match method.
//
// Values that implement this interface can be passed to [*Case.Rets] to control
// the matching algorithm for return values.
type Matcher interface {
	// Match reports whether a return value is considered a match. The argument
	// is of type RetValue so that it cannot be implemented accidentally.
	Match(RetValue) bool
}

// Any is a Matcher that matches any value.
var Any Matcher = anyMatcher{}

type anyMatcher struct{}

func (anyMatcher) Match(RetValue) bool { return true }

func match(matchers, actual []any) bool {
	for i, matcher := range matchers {
		if !matchOne(matcher, actual[i]) {
			return false
		}
	}
	return true
}

func matchOne(m, a any) bool {
	if m, ok := m.(Matcher); ok {
		return m.Match(a)
	}
	return reflect.DeepEqual(m, a)
}

func sprintArgs(args ...any) string {
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, arg)
	}
	return b.String()
}

func call(fn any, args []any) []any {
	argsReflect := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			// reflect.ValueOf(nil) returns a zero Value, but this is not what
			// we want. Work around this by taking the ValueOf a pointer to nil
			// and then get the Elem.
			// TODO(xiaq): This is now always using a nil value with type
			// interface{}. For more usability, inspect the type of fn to see
			// which type of nil this argument should be.
			var v any
			argsReflect[i] = reflect.ValueOf(&v).Elem()
		} else {
			argsReflect[i] = reflect.ValueOf(arg)
		}
	}
	retsReflect := reflect.ValueOf(fn).Call(argsReflect)
	rets := make([]any, len(retsReflect))
	for i, retReflect := range retsReflect {
		rets[i] = retReflect.Interface()
	}
	return rets
}

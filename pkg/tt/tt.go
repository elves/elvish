// Package tt supports table-driven tests with little boilerplate.
//
// See the test case for this package for example usage.
package tt

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// Table represents a test table.
type Table []*Case

// Case represents a test case. It is created by the C function, and offers
// setters that augment and return itself; those calls can be chained like
// C(...).Rets(...).
type Case struct {
	args         []any
	retsMatchers [][]any
}

// Args returns a new Case with the given arguments.
func Args(args ...any) *Case {
	return &Case{args: args}
}

// Rets modifies the test case so that it requires the return values to match
// the given values. It returns the receiver. The arguments may implement the
// Matcher interface, in which case its Match method is called with the actual
// return value. Otherwise, reflect.DeepEqual is used to determine matches.
func (c *Case) Rets(matchers ...any) *Case {
	c.retsMatchers = append(c.retsMatchers, matchers)
	return c
}

// FnToTest describes a function to test.
type FnToTest struct {
	name    string
	body    any
	argsFmt string
	retsFmt string
}

// Fn makes a new FnToTest with the given function name and body.
func Fn(name string, body any) *FnToTest {
	return &FnToTest{name: name, body: body}
}

// ArgsFmt sets the string for formatting arguments in test error messages, and
// return fn itself.
func (fn *FnToTest) ArgsFmt(s string) *FnToTest {
	fn.argsFmt = s
	return fn
}

// RetsFmt sets the string for formatting return values in test error messages,
// and return fn itself.
func (fn *FnToTest) RetsFmt(s string) *FnToTest {
	fn.retsFmt = s
	return fn
}

// T is the interface for accessing testing.T.
type T interface {
	Helper()
	Errorf(format string, args ...any)
}

// Test tests a function against test cases.
func Test(t T, fn *FnToTest, tests Table) {
	t.Helper()
	for _, test := range tests {
		rets := call(fn.body, test.args)
		for _, retsMatcher := range test.retsMatchers {
			if !match(retsMatcher, rets) {
				var args string
				if fn.argsFmt == "" {
					args = sprintArgs(test.args...)
				} else {
					args = fmt.Sprintf(fn.argsFmt, test.args...)
				}
				var diff string
				if len(retsMatcher) == 1 && len(rets) == 1 {
					diff = cmp.Diff(retsMatcher[0], rets[0], cmpopt)
				} else {
					diff = cmp.Diff(retsMatcher, rets, cmpopt)
				}
				t.Errorf("%s(%s) returns (-Wanted +Actual):\n%s", fn.name, args, diff)
			}
		}
	}
}

// RetValue is an empty interface used in the Matcher interface.
type RetValue any

// Matcher wraps the Match method.
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

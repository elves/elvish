// Package tt supports table-driven tests with little boilerplate.
//
// See the test case for this package for example usage.
package tt

import (
	"bytes"
	"fmt"
	"reflect"
)

// Table represents a test table.
type Table []*Case

// Case represents a test case. It is created by the C function, and offers
// setters that augment and return itself; those calls can be chained like
// C(...).Rets(...).
type Case struct {
	args         []interface{}
	retsMatchers [][]interface{}
}

// Args returns a new Case with the given arguments.
func Args(args ...interface{}) *Case {
	return &Case{args: args}
}

// Rets modifies the test case so that it requires the return values to match
// the given values. It returns the receiver. The arguments may implement the
// Matcher interface, in which case its Match method is called with the actual
// return value. Otherwise, reflect.DeepEqual is used to determine matches.
func (c *Case) Rets(matchers ...interface{}) *Case {
	c.retsMatchers = append(c.retsMatchers, matchers)
	return c
}

// FnToTest describes a function to test.
type FnToTest struct {
	name    string
	body    interface{}
	argsFmt string
	retsFmt string
}

// Fn makes a new FnToTest with the given function name and body.
func Fn(name string, body interface{}) *FnToTest {
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
	Errorf(format string, args ...interface{})
}

// Test tests a function against test cases.
func Test(t T, fn *FnToTest, tests Table) {
	t.Helper()
	for _, test := range tests {
		rets := call(fn.body, test.args)
		for _, retsMatcher := range test.retsMatchers {
			if !match(retsMatcher, rets) {
				var argsString, retsString, wantRetsString string
				if fn.argsFmt == "" {
					argsString = sprintArgs(test.args...)
				} else {
					argsString = fmt.Sprintf(fn.argsFmt, test.args...)
				}
				if fn.retsFmt == "" {
					retsString = sprintRets(rets...)
					wantRetsString = sprintRets(retsMatcher...)
				} else {
					retsString = fmt.Sprintf(fn.retsFmt, rets...)
					wantRetsString = fmt.Sprintf(fn.retsFmt, retsMatcher...)
				}
				t.Errorf("%s(%s) -> %s, want %s", fn.name, argsString, retsString, wantRetsString)
			}
		}
	}
}

// RetValue is an empty interface used in the Matcher interface.
type RetValue interface{}

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

func match(matchers, actual []interface{}) bool {
	for i, matcher := range matchers {
		if !matchOne(matcher, actual[i]) {
			return false
		}
	}
	return true
}

func matchOne(m, a interface{}) bool {
	if m, ok := m.(Matcher); ok {
		return m.Match(a)
	}
	return reflect.DeepEqual(m, a)
}

func sprintArgs(args ...interface{}) string {
	return sprintCommaDelimited(args...)
}

func sprintRets(rets ...interface{}) string {
	if len(rets) == 1 {
		return fmt.Sprint(rets[0])
	}
	return "(" + sprintCommaDelimited(rets...) + ")"
}

func sprintCommaDelimited(args ...interface{}) string {
	var b bytes.Buffer
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, arg)
	}
	return b.String()
}

func call(fn interface{}, args []interface{}) []interface{} {
	argsReflect := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			// reflect.ValueOf(nil) returns a zero Value, but this is not what
			// we want. Work around this by taking the ValueOf a pointer to nil
			// and then get the Elem.
			// TODO(xiaq): This is now always using a nil value with type
			// interface{}. For more usability, inspect the type of fn to see
			// which type of nil this argument should be.
			var v interface{}
			argsReflect[i] = reflect.ValueOf(&v).Elem()
		} else {
			argsReflect[i] = reflect.ValueOf(arg)
		}
	}
	retsReflect := reflect.ValueOf(fn).Call(argsReflect)
	rets := make([]interface{}, len(retsReflect))
	for i, retReflect := range retsReflect {
		rets[i] = retReflect.Interface()
	}
	return rets
}

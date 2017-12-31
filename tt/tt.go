// Package tt supports table-driven tests with little boilerplate.
//
// See the test case for this package for example usage.
package tt

import (
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

// C returns a new Case with the given arguments.
func C(args ...interface{}) *Case {
	return &Case{args: args}
}

// Rets modifies the test case so that it requires the return values to match
// the given values. It returns the receiver.
func (c *Case) Rets(matchers ...interface{}) *Case {
	c.retsMatchers = append(c.retsMatchers, matchers)
	return c
}

func match(matchers, actual []interface{}) bool {
	for i, matcher := range matchers {
		// TODO: Support custom matching strategy
		if !reflect.DeepEqual(matcher, actual[i]) {
			return false
		}
	}
	return true
}

// Fn describes a function under test.
type Fn struct {
	Name   string
	ArgFmt string
	RetFmt string
	Body   interface{}
}

// T is the interface for accessing testing.T.
type T interface {
	Errorf(format string, args ...interface{})
}

// Test tests a function against test cases.
func Test(t T, fn Fn, tests Table) {
	for _, test := range tests {
		rets := call(fn.Body, test.args)
		for _, retsMatcher := range test.retsMatchers {
			if !match(retsMatcher, rets) {
				argsString := fmt.Sprintf(fn.ArgFmt, test.args...)
				retsString := fmt.Sprintf(fn.RetFmt, rets...)
				wantRetsString := fmt.Sprintf(fn.RetFmt, retsMatcher...)
				t.Errorf("%s%s -> %s, want %s", fn.Name, argsString, retsString, wantRetsString)
			}
		}
	}
}

func call(fn interface{}, args []interface{}) []interface{} {
	argsReflect := make([]reflect.Value, len(args))
	for i, arg := range args {
		argsReflect[i] = reflect.ValueOf(arg)
	}
	retsReflect := reflect.ValueOf(fn).Call(argsReflect)
	rets := make([]interface{}, len(retsReflect))
	for i, retReflect := range retsReflect {
		rets[i] = retReflect.Interface()
	}
	return rets
}

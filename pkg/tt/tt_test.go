package tt

import (
	"fmt"
	"strings"
	"testing"
)

// Simple functions to test.

func add(x, y int) int {
	return x + y
}

func addsub(x int, y int) (int, int) {
	return x + y, x - y
}

func TestTest(t *testing.T) {
	Test(t, test,
		It("reports no errors for passing tests").
			Args(add, Args(1, 1).Rets(2)).
			Rets([]testResult{{"", nil}}),
		It("supports multiple tests").
			Args(add, Args(1, 1).Rets(2), Args(1, 2).Rets(3)),
		It("supports for multiple return values").
			Args(addsub, Args(1, 2).Rets(3, -1)).
			Rets([]testResult{{"", nil}}),
		It("supports named tests").
			Args(add, It("can add 1 and 1").Args(1, 1).Rets(2)).
			Rets([]testResult{{"can add 1 and 1", nil}}),

		It("reports error for failed test").
			Args(Fn(add).Named("add"), Args(2, 2).Rets(5)).
			Rets(testResultsMatcher{
				{"", []string{"add(2, 2) returns (-want +got):\n"}},
			}),
		It("respects custom argument format strings when reporting errors").
			Args(Fn(add).Named("add").ArgsFmt("x = %d, y = %d"), Args(1, 2).Rets(5)).
			Rets(testResultsMatcher{
				{"", []string{"add(x = 1, y = 2) returns (-want +got):\n"}},
			}),
	)
}

// An alternative to the exported [Test] that uses a mock test runner that
// collects results from all the subtests.
func test(fn any, tests ...*Case) []testResult {
	var tr mockTestRunner
	testInner[*mockSubtestRunner](&tr, fn, tests...)
	return tr
}

// Mock implementations of testRunner and subtestRunner.

type testResult struct {
	Name   string
	Errors []string
}

type mockTestRunner []testResult

func (tr *mockTestRunner) Helper() {}

func (tr *mockTestRunner) Run(name string, f func(*mockSubtestRunner)) bool {
	sr := mockSubtestRunner{name, nil}
	f(&sr)
	*tr = append(*tr, testResult(sr))
	return len(sr.Errors) == 0
}

type mockSubtestRunner testResult

func (sr *mockSubtestRunner) Errorf(format string, args ...any) {
	sr.Errors = append(sr.Errors, fmt.Sprintf(format, args...))
}

// Matches []testResult, but doesn't check the exact content of the error
// messages, only that they contain a substring.
type testResultsMatcher []testResult

func (m testResultsMatcher) Match(ret RetValue) bool {
	results, ok := ret.([]testResult)
	if !ok {
		return false
	}
	if len(results) != len(m) {
		return false
	}
	for i, result := range results {
		if result.Name != m[i].Name || len(result.Errors) != len(m[i].Errors) {
			return false
		}
		for i, s := range result.Errors {
			if !strings.Contains(s, m[i].Errors[i]) {
				return false
			}
		}
	}
	return true
}

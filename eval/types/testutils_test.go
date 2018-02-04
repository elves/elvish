package types

import "github.com/elves/elvish/tt"

// anyType matches anything.
type anyType struct{}

var any = anyType{}

func (anyType) Match(tt.RetValue) bool { return true }

// anyErrorType matches any value that satisfies the error interface.
type anyErrorType struct{}

var anyError = anyErrorType{}

func (e anyErrorType) Match(a tt.RetValue) bool {
	_, ok := a.(error)
	return ok
}

// equalMatcher matches the return value using Equal.
type equalMatcher struct {
	r interface{}
}

func eq(r interface{}) tt.Matcher { return equalMatcher{r} }

func (em equalMatcher) Match(a tt.RetValue) bool { return Equal(em.r, a) }

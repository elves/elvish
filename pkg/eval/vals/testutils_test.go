package vals

import "github.com/elves/elvish/pkg/tt"

var Args = tt.Args

// anyType matches anything.
type anyType struct{}

var any = anyType{}

func (anyType) Match(tt.RetValue) bool { return true }

// equalMatcher matches the return value using Equal.
type equalMatcher struct {
	r interface{}
}

func eq(r interface{}) tt.Matcher { return equalMatcher{r} }

func (em equalMatcher) Match(a tt.RetValue) bool { return Equal(em.r, a) }

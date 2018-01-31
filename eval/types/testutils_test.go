package types

import "github.com/elves/elvish/tt"

type equalMatcher struct {
	r interface{}
}

func eq(r interface{}) tt.Matcher {
	return equalMatcher{r}
}

func (em equalMatcher) Match(a tt.RetValue) bool {
	return Equal(em.r, a)
}

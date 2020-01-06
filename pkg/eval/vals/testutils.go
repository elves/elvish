package vals

import "github.com/elves/elvish/pkg/tt"

// Eq returns a tt.Matcher that matches using the Equal function.
func Eq(r interface{}) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want interface{} }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }

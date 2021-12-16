package vals

import "src.elv.sh/pkg/tt"

// Returns a tt.Matcher that matches using the Equal function.
func eq(r interface{}) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want interface{} }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }

package vals

import (
	"src.elv.sh/pkg/tt"
)

// Returns a tt.Matcher that matches using the Equal function.
func eq(r any) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want any }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }

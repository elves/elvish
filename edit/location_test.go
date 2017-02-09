package edit

import (
	"testing"

	"github.com/elves/elvish/store"
)

var (
	theLocation = newLocation([]store.Dir{
		{"/src/github.com/elves/elvish", 300},
		{"/src/home/xyz", 233},
		{"/usr/elves/elvish", 6},
	})

	locationFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"233", unstyled("/src/home/xyz")},
			{"6", unstyled("/usr/elves/elvish")}}},
		{"/s", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"233", unstyled("/src/home/xyz")},
			{"6", unstyled("/usr/elves/elvish")}}},
		{"/e/e", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"6", unstyled("/usr/elves/elvish")}}},
		{"x", []shown{{"233", unstyled("/src/home/xyz")}}},
	}
)

func TestLocation(t *testing.T) {
	testListingFilter(t, "theLocation", theLocation, locationFilterTests)
}

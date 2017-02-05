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
		{"", []styled{
			unstyled(" 300 /src/github.com/elves/elvish"),
			unstyled(" 233 /src/home/xyz"),
			unstyled("   6 /usr/elves/elvish")}},
		{"/s", []styled{
			unstyled(" 300 /src/github.com/elves/elvish"),
			unstyled(" 233 /src/home/xyz"),
			unstyled("   6 /usr/elves/elvish")}},
		{"/e/e", []styled{
			unstyled(" 300 /src/github.com/elves/elvish"),
			unstyled("   6 /usr/elves/elvish")}},
		{"x", []styled{unstyled(" 233 /src/home/xyz")}},
	}
)

func TestLocation(t *testing.T) {
	testListingFilter(t, "theLocation", theLocation, locationFilterTests)
}

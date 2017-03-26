package edit

import (
	"testing"

	"github.com/elves/elvish/store"
)

var (
	theLocation = newLocation([]store.Dir{
		{"/src/github.com/elves/elvish", 300},
		{"/src/home/xyz", 233},
		{"/home/dir", 100},
		{"/foo/\nbar", 77},
		{"/usr/elves/elvish", 6},
	}, "/home")

	locationFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"233", unstyled("/src/home/xyz")},
			{"100", unstyled("~/dir")},       // home is abbreviated
			{"77", unstyled(`"/foo/\nbar"`)}, // special char is quoted
			{"6", unstyled("/usr/elves/elvish")}}},
		{"/s", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"233", unstyled("/src/home/xyz")},
			{"6", unstyled("/usr/elves/elvish")}}},
		{"/e/e", []shown{
			{"300", unstyled("/src/github.com/elves/elvish")},
			{"6", unstyled("/usr/elves/elvish")}}},
		{"x", []shown{{"233", unstyled("/src/home/xyz")}}},
		// Matchers operate on the displayed text, not the actual path.
		// 1. Home directory is abbreviated to ~, and is matched by ~, but not by
		//    the actual path.
		{"~", []shown{{"100", unstyled("~/dir")}}},
		{"home", []shown{{"233", unstyled("/src/home/xyz")}}},
		// 2. Special characters are quoted, and are matched by the quoted form,
		//    not by the actual form.
		{"\n", []shown{}},
		{"\\n", []shown{{"77", unstyled(`"/foo/\nbar"`)}}},
	}
)

func TestLocation(t *testing.T) {
	testListingFilter(t, "theLocation", theLocation, locationFilterTests)
}

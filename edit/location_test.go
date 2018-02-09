package edit

import (
	"testing"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
)

var (
	theLocation = newLocation([]storedefs.Dir{
		{"/pinned", pinnedScore},
		{"/src/github.com/elves/elvish", 300},
		{"/src/home/xyz", 233},
		{"/home/dir", 100},
		{"/foo/\nbar", 77},
		{"/usr/elves/elvish", 6},
	}, "/home")

	locationFilterTests = []eddefs.ListingProviderFilterTest{
		{"", []eddefs.ListingShown{
			{"*", ui.Unstyled("/pinned")},
			{"300", ui.Unstyled("/src/github.com/elves/elvish")},
			{"233", ui.Unstyled("/src/home/xyz")},
			{"100", ui.Unstyled("~/dir")},       // home is abbreviated
			{"77", ui.Unstyled(`"/foo/\nbar"`)}, // special char is quoted
			{"6", ui.Unstyled("/usr/elves/elvish")}}},
		{"/s", []eddefs.ListingShown{
			{"300", ui.Unstyled("/src/github.com/elves/elvish")},
			{"233", ui.Unstyled("/src/home/xyz")},
			{"6", ui.Unstyled("/usr/elves/elvish")}}},
		{"/e/e", []eddefs.ListingShown{
			{"300", ui.Unstyled("/src/github.com/elves/elvish")},
			{"6", ui.Unstyled("/usr/elves/elvish")}}},
		{"x", []eddefs.ListingShown{{"233", ui.Unstyled("/src/home/xyz")}}},
		// Matchers operate on the displayed text, not the actual path.
		// 1. Home directory is abbreviated to ~, and is matched by ~, but not by
		//    the actual path.
		{"~", []eddefs.ListingShown{{"100", ui.Unstyled("~/dir")}}},
		{"home", []eddefs.ListingShown{{"233", ui.Unstyled("/src/home/xyz")}}},
		// 2. Special characters are quoted, and are matched by the quoted form,
		//    not by the actual form.
		{"\n", []eddefs.ListingShown{}},
		{"\\n", []eddefs.ListingShown{{"77", ui.Unstyled(`"/foo/\nbar"`)}}},
	}
)

func TestLocation(t *testing.T) {
	eddefs.TestListingProviderFilter(t, "theLocation", theLocation, locationFilterTests)
}

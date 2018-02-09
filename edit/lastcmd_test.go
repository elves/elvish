package edit

import (
	"testing"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
)

var (
	theLine    = "qw search 'foo bar ~y'"
	theLastCmd = newLastCmd(theLine)

	lastcmdFilterTests = []eddefs.ListingProviderFilterTest{
		{"", []eddefs.ListingShown{
			{"M-1", ui.Unstyled(theLine)},
			{"0", ui.Unstyled("qw")},
			{"1", ui.Unstyled("search")},
			{"2", ui.Unstyled("'foo bar ~y'")}}},
		{"1", []eddefs.ListingShown{{"1", ui.Unstyled("search")}}},
		{"-", []eddefs.ListingShown{
			{"M-1", ui.Unstyled(theLine)},
			{"-3", ui.Unstyled("qw")},
			{"-2", ui.Unstyled("search")},
			{"-1", ui.Unstyled("'foo bar ~y'")}}},
		{"-1", []eddefs.ListingShown{{"-1", ui.Unstyled("'foo bar ~y'")}}},
	}
)

func TestLastCmd(t *testing.T) {
	eddefs.TestListingProviderFilter(t, "theLastCmd", theLastCmd, lastcmdFilterTests)
}

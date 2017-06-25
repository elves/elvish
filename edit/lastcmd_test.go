package edit

import (
	"testing"

	"github.com/elves/elvish/edit/ui"
)

var (
	theLine    = "qw search 'foo bar ~y'"
	theLastCmd = newLastCmd(theLine)

	lastcmdFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"M-,", ui.Unstyled(theLine)},
			{"0", ui.Unstyled("qw")},
			{"1", ui.Unstyled("search")},
			{"2", ui.Unstyled("'foo bar ~y'")}}},
		{"1", []shown{{"1", ui.Unstyled("search")}}},
		{"-", []shown{
			{"M-,", ui.Unstyled(theLine)},
			{"-3", ui.Unstyled("qw")},
			{"-2", ui.Unstyled("search")},
			{"-1", ui.Unstyled("'foo bar ~y'")}}},
		{"-1", []shown{{"-1", ui.Unstyled("'foo bar ~y'")}}},
	}
)

func TestLastCmd(t *testing.T) {
	testListingFilter(t, "theLastCmd", theLastCmd, lastcmdFilterTests)
}

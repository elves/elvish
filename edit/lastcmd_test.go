package edit

import (
	"testing"

	"github.com/elves/elvish/edit/listing"
	"github.com/elves/elvish/edit/ui"
)

var (
	theLine    = "qw search 'foo bar ~y'"
	theLastCmd = newLastCmd(theLine)

	lastcmdFilterTests = []listing.FilterTest{
		{"", []listing.Shown{
			{"M-1", ui.Unstyled(theLine)},
			{"0", ui.Unstyled("qw")},
			{"1", ui.Unstyled("search")},
			{"2", ui.Unstyled("'foo bar ~y'")}}},
		{"1", []listing.Shown{{"1", ui.Unstyled("search")}}},
		{"-", []listing.Shown{
			{"M-1", ui.Unstyled(theLine)},
			{"-3", ui.Unstyled("qw")},
			{"-2", ui.Unstyled("search")},
			{"-1", ui.Unstyled("'foo bar ~y'")}}},
		{"-1", []listing.Shown{{"-1", ui.Unstyled("'foo bar ~y'")}}},
	}
)

func TestLastCmd(t *testing.T) {
	listing.TestProviderFilter(t, "theLastCmd", theLastCmd, lastcmdFilterTests)
}

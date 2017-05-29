package edit

import (
	"testing"

	"github.com/elves/elvish/edit/ui"
)

var (
	theHistList = newHistlist([]string{"ls", "echo lalala", "ls"})

	histlistFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"0", ui.Unstyled("ls")},
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
		{"l", []shown{
			{"0", ui.Unstyled("ls")},
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
		// {"ch", []styled{unstyled("1 echo lalala")}},
	}
)

func TestHistlist(t *testing.T) {
	testListingFilter(t, "theHistList", theHistList, histlistFilterTests)
}

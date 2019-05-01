package history

import (
	"testing"

	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
)

var (
	theHistList = newHistlist(
		[]histutil.Entry{{"ls", 1}, {"echo lalala", 2}, {"ls", 3}})

	histlistDedupFilterTests = []eddefs.ListingProviderFilterTest{
		{"", []eddefs.ListingShown{
			{"2", ui.Unstyled("echo lalala")},
			{"3", ui.Unstyled("ls")}}},
		{"l", []eddefs.ListingShown{
			{"2", ui.Unstyled("echo lalala")},
			{"3", ui.Unstyled("ls")}}},
	}

	histlistNoDedupFilterTests = []eddefs.ListingProviderFilterTest{
		{"", []eddefs.ListingShown{
			{"1", ui.Unstyled("ls")},
			{"2", ui.Unstyled("echo lalala")},
			{"3", ui.Unstyled("ls")}}},
		{"l", []eddefs.ListingShown{
			{"1", ui.Unstyled("ls")},
			{"2", ui.Unstyled("echo lalala")},
			{"3", ui.Unstyled("ls")}}},
	}
)

func TestHistlist(t *testing.T) {
	eddefs.TestListingProviderFilter(t, "theHistList", theHistList, histlistDedupFilterTests)
	theHistList.dedup = false
	eddefs.TestListingProviderFilter(t, "theHistList", theHistList, histlistNoDedupFilterTests)
}

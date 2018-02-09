package edit

import (
	"testing"

	"github.com/elves/elvish/edit/listing"
	"github.com/elves/elvish/edit/ui"
)

var (
	theHistList = newHistlist([]string{"ls", "echo lalala", "ls"})

	histlistDedupFilterTests = []listing.FilterTest{
		{"", []listing.Shown{
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
		{"l", []listing.Shown{
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
	}

	histlistNoDedupFilterTests = []listing.FilterTest{
		{"", []listing.Shown{
			{"0", ui.Unstyled("ls")},
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
		{"l", []listing.Shown{
			{"0", ui.Unstyled("ls")},
			{"1", ui.Unstyled("echo lalala")},
			{"2", ui.Unstyled("ls")}}},
	}
)

func TestHistlist(t *testing.T) {
	listing.TestProviderFilter(t, "theHistList", theHistList, histlistDedupFilterTests)
	theHistList.dedup = false
	listing.TestProviderFilter(t, "theHistList", theHistList, histlistNoDedupFilterTests)
}

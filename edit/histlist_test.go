package edit

import "testing"

var (
	theHistList = newHistlist([]string{"ls", "echo lalala", "ls"})

	histlistFilterTests = []listingFilterTestCases{
		{"", []styled{
			unstyled("0 ls"),
			unstyled("1 echo lalala"),
			unstyled("2 ls")}},
		{"l", []styled{
			unstyled("0 ls"),
			unstyled("1 echo lalala"),
			unstyled("2 ls")}},
		// {"ch", []styled{unstyled("1 echo lalala")}},
	}
)

func TestHistlist(t *testing.T) {
	testListingFilter(t, "theHistList", theHistList, histlistFilterTests)
}

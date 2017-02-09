package edit

import "testing"

var (
	theHistList = newHistlist([]string{"ls", "echo lalala", "ls"})

	histlistFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"0", unstyled("ls")},
			{"1", unstyled("echo lalala")},
			{"2", unstyled("ls")}}},
		{"l", []shown{
			{"0", unstyled("ls")},
			{"1", unstyled("echo lalala")},
			{"2", unstyled("ls")}}},
		// {"ch", []styled{unstyled("1 echo lalala")}},
	}
)

func TestHistlist(t *testing.T) {
	testListingFilter(t, "theHistList", theHistList, histlistFilterTests)
}

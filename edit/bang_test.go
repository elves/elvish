package edit

import "testing"

var (
	theLine = "qw search 'foo bar ~y'"
	theBang = newBang(theLine)

	bangFilterTests = []listingFilterTestCases{
		{"", []styled{
			unstyled("M-, " + theLine),
			unstyled("  0 qw"),
			unstyled("  1 search"),
			unstyled("  2 'foo bar ~y'")}},
		{"1", []styled{unstyled("  1 search")}},
		{"-", []styled{
			unstyled("M-, " + theLine),
			unstyled(" -3 qw"),
			unstyled(" -2 search"),
			unstyled(" -1 'foo bar ~y'")}},
		{"-1", []styled{unstyled(" -1 'foo bar ~y'")}},
	}
)

func TestBang(t *testing.T) {
	testListingFilter(t, "theBang", theBang, bangFilterTests)
}

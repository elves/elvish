package edit

import "testing"

var (
	theLine = "qw search 'foo bar ~y'"
	theBang = newBang(theLine)

	bangFilterTests = []listingFilterTestCases{
		{"", []shown{
			{"M-,", unstyled(theLine)},
			{"0", unstyled("qw")},
			{"1", unstyled("search")},
			{"2", unstyled("'foo bar ~y'")}}},
		{"1", []shown{{"1", unstyled("search")}}},
		{"-", []shown{
			{"M-,", unstyled(theLine)},
			{"-3", unstyled("qw")},
			{"-2", unstyled("search")},
			{"-1", unstyled("'foo bar ~y'")}}},
		{"-1", []shown{{"-1", unstyled("'foo bar ~y'")}}},
	}
)

func TestBang(t *testing.T) {
	testListingFilter(t, "theBang", theBang, bangFilterTests)
}

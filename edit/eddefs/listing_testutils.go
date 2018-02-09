package eddefs

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
)

type ListingShown struct {
	Header  string
	Content ui.Styled
}

type ListingProviderFilterTest struct {
	Filter     string
	WantShowns []ListingShown
}

func TestListingProviderFilter(t *testing.T, name string, ls ListingProvider, testcases []ListingProviderFilterTest) {
	for _, testcase := range testcases {
		ls.Filter(testcase.Filter)

		l := ls.Len()
		if l != len(testcase.WantShowns) {
			t.Errorf("%s.Len() -> %d, want %d (filter was %q)",
				name, l, len(testcase.WantShowns), testcase.Filter)
		} else {
			for i, want := range testcase.WantShowns {
				header, content := ls.Show(i)
				if header != want.Header || !reflect.DeepEqual(content, want.Content) {
					t.Errorf("%s.Show(%d) => (%v, %v), want (%v, %v) (filter was %q)",
						name, i, header, content, want.Header, want.Content, testcase.Filter)
				}
			}
		}
	}
}

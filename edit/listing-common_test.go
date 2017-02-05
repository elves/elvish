package edit

import (
	"reflect"
	"testing"
)

type listingFilterTestCases struct {
	filter     string
	wantShowns []styled
}

func testListingFilter(t *testing.T, name string, ls listingProvider, testcases []listingFilterTestCases) {
	for _, testcase := range testcases {
		ls.Filter(testcase.filter)

		l := ls.Len()
		if l != len(testcase.wantShowns) {
			t.Errorf("%s.Len() -> %d, want %d (filter was %q)",
				name, l, len(testcase.wantShowns), testcase.filter)
		}
		for i, want := range testcase.wantShowns {
			shown := ls.Show(i)
			if !reflect.DeepEqual(shown, want) {
				t.Errorf("%s.Show(%d) => %v, want %v (filter was %q)",
					name, i, shown, want, testcase.filter)
			}
		}
	}
}

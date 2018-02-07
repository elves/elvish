package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
)

type shown struct {
	header  string
	content ui.Styled
}

type listingFilterTestCases struct {
	filter     string
	wantShowns []shown
}

func testListingFilter(t *testing.T, name string, l *listingState, testcases []listingFilterTestCases) {
	ls := l.provider
	for _, testcase := range testcases {
		ls.Filter(testcase.filter)

		l := ls.Len()
		if l != len(testcase.wantShowns) {
			t.Errorf("%s.Len() -> %d, want %d (filter was %q)",
				name, l, len(testcase.wantShowns), testcase.filter)
		} else {
			for i, want := range testcase.wantShowns {
				header, content := ls.Show(i)
				if header != want.header || !reflect.DeepEqual(content, want.content) {
					t.Errorf("%s.Show(%d) => (%v, %v), want (%v, %v) (filter was %q)",
						name, i, header, content, want.header, want.content, testcase.filter)
				}
			}
		}
	}
}

package listing

import (
	"reflect"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

type itemsMatcher struct{ wantShown []styled.Text }

func (m itemsMatcher) Match(v tt.RetValue) bool {
	items := v.(Items)
	if items.Len() != len(m.wantShown) {
		return false
	}
	for i, wantShown := range m.wantShown {
		if !reflect.DeepEqual(wantShown, items.Show(i)) {
			return false
		}
	}
	return true
}

// MatchItems returns a Matcher that checks if all of the entries produced in a
// given Items match what is expected.
func MatchItems(wantShown ...styled.Text) tt.Matcher {
	return itemsMatcher{wantShown}
}

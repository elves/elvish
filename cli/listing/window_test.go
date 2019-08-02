package listing

import (
	"testing"

	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestFindWindow(t *testing.T) {
	tt.Test(t, tt.Fn("findWindow", findWindow), tt.Table{
		// selected = 0: always show a widow starting from 0, regardless of
		// the value of oldFirst
		Args(fakeItems{10}, 0, 0, 6).Rets(0, 0),
		Args(fakeItems{10}, 1, 0, 6).Rets(0, 0),
		// selected = n-1: always show a window ending at n-1, regardless of the
		// value of oldFirst
		Args(fakeItems{10}, 0, 9, 6).Rets(4, 0),
		Args(fakeItems{10}, 8, 9, 6).Rets(4, 0),
		// selected = 3, oldFirst = 2 (likely because previous selected = 4).
		// Adjust first -> 1 to satisfy the upward respect distance of 2.
		Args(fakeItems{10}, 2, 3, 6).Rets(1, 0),
		// selected = 6, oldFirst = 2 (likely because previous selected = 7).
		// Adjust first -> 3 to satisfy the downward respect distance of 2.
		Args(fakeItems{10}, 2, 6, 6).Rets(3, 0),

		// There is not enough budget to achieve respect distance on both sides.
		// Split the budget in half.
		Args(fakeItems{10}, 1, 3, 3).Rets(2, 0),
		Args(fakeItems{10}, 0, 3, 3).Rets(2, 0),

		// There is just enough distance to fit the selected item. Only show the
		// selected item.
		Args(fakeItems{10}, 0, 2, 1).Rets(2, 0),
	})
}

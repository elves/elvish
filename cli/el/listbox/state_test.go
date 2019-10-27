package listbox

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMakeState(t *testing.T) {
	tt.Test(t, tt.Fn("MakeState", MakeState), tt.Table{
		tt.Args(TestItems{NItems: 10}, false).
			Rets(State{Items: TestItems{NItems: 10}, Selected: 0, First: 0}),
		tt.Args(TestItems{NItems: 10}, true).
			Rets(State{Items: TestItems{NItems: 10}, Selected: 9, First: 0}),
	})
}

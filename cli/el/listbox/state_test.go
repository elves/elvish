package listbox

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMakeState(t *testing.T) {
	tt.Test(t, tt.Fn("MakeState", MakeState), tt.Table{
		tt.Args(TestItems{NItems: 10}, false).Rets(State{TestItems{NItems: 10}, 0, 0}),
		tt.Args(TestItems{NItems: 10}, true).Rets(State{TestItems{NItems: 10}, 9, 0}),
	})
}

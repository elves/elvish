package listbox

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMakeState(t *testing.T) {
	tt.Test(t, tt.Fn("MakeState", MakeState), tt.Table{
		tt.Args(TestItemer{}, 10, false).Rets(State{TestItemer{}, 10, 0, 0}),
		tt.Args(TestItemer{}, 10, true).Rets(State{TestItemer{}, 10, 9, 0}),
	})
}

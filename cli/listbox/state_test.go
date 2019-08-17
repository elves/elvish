package listbox

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMakeState(t *testing.T) {
	tt.Test(t, tt.Fn("MakeState", MakeState), tt.Table{
		tt.Args(itemer{}, 10, false).Rets(State{itemer{}, 10, 0, 0}),
		tt.Args(itemer{}, 10, true).Rets(State{itemer{}, 10, 9, 0}),
	})
}

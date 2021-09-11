package vals

import (
	"testing"

	"src.elv.sh/pkg/testutil"
	. "src.elv.sh/pkg/tt"
)

func TestMakeMap_PanicsWithOddNumberOfArguments(t *testing.T) {
	Test(t, Fn("Recover", testutil.Recover), Table{
		Args(func() { MakeMap("foo") }).Rets("odd number of arguments to MakeMap"),
	})
}

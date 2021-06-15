package testutil

import (
	"testing"

	. "src.elv.sh/pkg/tt"
)

func TestRecover(t *testing.T) {
	Test(t, Fn("Recover", Recover), Table{
		Args(func() {}).Rets(nil),
		Args(func() {
			panic("unreachable")
		}).Rets("unreachable"),
	})
}

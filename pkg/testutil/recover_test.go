package testutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestRecover(t *testing.T) {
	tt.Test(t, tt.Fn("Recover", Recover), tt.Table{
		tt.Args(func() {}).Rets(nil),
		tt.Args(func() {
			panic("unreachable")
		}).Rets("unreachable"),
	})
}

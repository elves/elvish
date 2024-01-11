package testutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestRecover(t *testing.T) {
	tt.Test(t, Recover,
		Args(func() {}).Rets(nil),
		Args(func() {
			panic("unreachable")
		}).Rets("unreachable"),
	)
}

package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestErrors(t *testing.T) {
	tt.Test(t, error.Error,
		Args(cannotIterate{"num"}).Rets("cannot iterate num"),
		Args(cannotIterateKeysOf{"num"}).Rets("cannot iterate keys of num"),
	)
}

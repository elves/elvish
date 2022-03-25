package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestErrors(t *testing.T) {
	tt.Test(t, tt.Fn("error.Error", error.Error), tt.Table{
		Args(cannotIterate{"num"}).Rets("cannot iterate num"),
		Args(cannotIterateKeysOf{"num"}).Rets("cannot iterate keys of num"),
	})
}

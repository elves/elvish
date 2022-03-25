package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestLen(t *testing.T) {
	tt.Test(t, tt.Fn("Len", Len), tt.Table{
		tt.Args("foobar").Rets(6),
		tt.Args(10).Rets(-1),
	})
}

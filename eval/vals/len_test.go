package vals

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestLen(t *testing.T) {
	tt.Test(t, tt.Fn("Len", Len), tt.Table{
		Args("foobar").Rets(6),
		Args(testStructMap{}).Rets(2),
		Args(10).Rets(-1),
	})
}

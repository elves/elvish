package vals

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

func TestLen(t *testing.T) {
	Test(t, Fn("Len", Len), Table{
		Args("foobar").Rets(6),
		Args(testStructMap{}).Rets(2),
		Args(10).Rets(-1),
	})
}

package vals

import (
	"testing"

	. "src.elv.sh/pkg/tt"
)

func TestLen(t *testing.T) {
	Test(t, Fn("Len", Len), Table{
		Args("foobar").Rets(6),
		Args(10).Rets(-1),
	})
}

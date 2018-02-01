package eval

import (
	"testing"

	"github.com/elves/elvish/tt"
)

var splitQNameTests = tt.Table{
	tt.Args("a").Rets([]string{"a"}),
	tt.Args("a:b").Rets([]string{"a:", "b"}),
	tt.Args("a:b:").Rets([]string{"a:", "b:"}),
	tt.Args("a:b:c:d").Rets([]string{"a:", "b:", "c:", "d"}),
}

func TestSplitQName(t *testing.T) {
	tt.Test(t, tt.Fn("splitQName", splitQName), splitQNameTests)
}

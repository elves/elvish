package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestSplitSigil(t *testing.T) {
	tt.Test(t, tt.Fn("SplitSigil", SplitSigil), tt.Table{
		Args("").Rets("", ""),
		Args("x").Rets("", "x"),
		Args("@x").Rets("@", "x"),
		Args("a:b").Rets("", "a:b"),
		Args("@a:b").Rets("@", "a:b"),
	})
}

func TestSplitQName(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQName", SplitQName), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("a", ""),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:", "b:"),
		Args("a:b:c").Rets("a:", "b:c"),
		Args("a:b:c:").Rets("a:", "b:c:"),
	})
}

func TestSplitQNameSegs(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQNameSegs", SplitQNameSegs), tt.Table{
		Args("").Rets([]string{}),
		Args("a").Rets([]string{"a"}),
		Args("a:").Rets([]string{"a:"}),
		Args("a:b").Rets([]string{"a:", "b"}),
		Args("a:b:").Rets([]string{"a:", "b:"}),
		Args("a:b:c").Rets([]string{"a:", "b:", "c"}),
		Args("a:b:c:").Rets([]string{"a:", "b:", "c:"}),
	})
}

func TestSplitIncompleteQNameNs(t *testing.T) {
	tt.Test(t, tt.Fn("SplitIncompleteQNameNs", SplitIncompleteQNameNs), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:b:", ""),
		Args("a:b:c").Rets("a:b:", "c"),
		Args("a:b:c:").Rets("a:b:c:", ""),
	})
}

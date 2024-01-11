package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"

	"src.elv.sh/pkg/tt"
)

func TestSplitSigil(t *testing.T) {
	tt.Test(t, SplitSigil,
		Args("").Rets("", ""),
		Args("x").Rets("", "x"),
		Args("@x").Rets("@", "x"),
		Args("a:b").Rets("", "a:b"),
		Args("@a:b").Rets("@", "a:b"),
	)
}

func TestSplitQName(t *testing.T) {
	tt.Test(t, SplitQName,
		Args("").Rets("", ""),
		Args("a").Rets("a", ""),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:", "b:"),
		Args("a:b:c").Rets("a:", "b:c"),
		Args("a:b:c:").Rets("a:", "b:c:"),
	)
}

func TestSplitQNameSegs(t *testing.T) {
	tt.Test(t, SplitQNameSegs,
		Args("").Rets([]string{}),
		Args("a").Rets([]string{"a"}),
		Args("a:").Rets([]string{"a:"}),
		Args("a:b").Rets([]string{"a:", "b"}),
		Args("a:b:").Rets([]string{"a:", "b:"}),
		Args("a:b:c").Rets([]string{"a:", "b:", "c"}),
		Args("a:b:c:").Rets([]string{"a:", "b:", "c:"}),
	)
}

func TestSplitIncompleteQNameNs(t *testing.T) {
	tt.Test(t, SplitIncompleteQNameNs,
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:b:", ""),
		Args("a:b:c").Rets("a:b:", "c"),
		Args("a:b:c:").Rets("a:b:c:", ""),
	)
}

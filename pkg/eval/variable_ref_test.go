package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

var Args = tt.Args

func TestSplitVariableRef(t *testing.T) {
	tt.Test(t, tt.Fn("SplitVariableRef", SplitVariableRef), tt.Table{
		Args("").Rets("", ""),
		Args("x").Rets("", "x"),
		Args("@x").Rets("@", "x"),
		Args("a:b").Rets("", "a:b"),
		Args("@a:b").Rets("@", "a:b"),
	})
}

func TestSplitQNameNs(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQNameNs", SplitQNameNs), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("", "a:"),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:", "b:"),
		Args("a:b:c").Rets("a:b:", "c"),
		Args("a:b:c:").Rets("a:b:", "c:"),
	})
}

func TestSplitQNameNsIncomplete(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQNameNsIncomplete", SplitQNameNsIncomplete), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:b:", ""),
		Args("a:b:c").Rets("a:b:", "c"),
		Args("a:b:c:").Rets("a:b:c:", ""),
	})
}

func TestSplitQNameNsFirst(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQNameNsFirst", SplitQNameNsFirst), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("", "a:"),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:", "b:"),
		Args("a:b:c").Rets("a:", "b:c"),
		Args("a:b:c:").Rets("a:", "b:c:"),
	})
}

func TestSplitIncompleteQNameFirstNs(t *testing.T) {
	tt.Test(t, tt.Fn("SplitIncompleteQNameFirstNs", SplitIncompleteQNameFirstNs), tt.Table{
		Args("").Rets("", ""),
		Args("a").Rets("", "a"),
		Args("a:").Rets("a:", ""),
		Args("a:b").Rets("a:", "b"),
		Args("a:b:").Rets("a:", "b:"),
		Args("a:b:c").Rets("a:", "b:c"),
		Args("a:b:c:").Rets("a:", "b:c:"),
	})
}

func TestSplitQNameNsSegs(t *testing.T) {
	tt.Test(t, tt.Fn("SplitQNameNsSegs", SplitQNameNsSegs), tt.Table{
		Args("").Rets([]string{}),
		Args("a").Rets([]string{"a"}),
		Args("a:").Rets([]string{"a:"}),
		Args("a:b").Rets([]string{"a:", "b"}),
		Args("a:b:").Rets([]string{"a:", "b:"}),
		Args("a:b:c").Rets([]string{"a:", "b:", "c"}),
		Args("a:b:c:").Rets([]string{"a:", "b:", "c:"}),
	})
}

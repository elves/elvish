package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

// An implementation of Iterator.
type iterator struct{ elements []any }

func (i iterator) Iterate(f func(any) bool) {
	Feed(f, i.elements...)
}

// A non-implementation of Iterator.
type nonIterator struct{}

func TestCanIterate(t *testing.T) {
	tt.Test(t, tt.Fn("CanIterate", CanIterate), tt.Table{
		tt.Args("foo").Rets(true),
		tt.Args(MakeList("foo", "bar")).Rets(true),
		tt.Args(iterator{vs("a", "b")}).Rets(true),
		tt.Args(nonIterator{}).Rets(false),
	})
}

func TestCollect(t *testing.T) {
	tt.Test(t, tt.Fn("Collect", Collect), tt.Table{
		tt.Args("foo").Rets(vs("f", "o", "o"), nil),
		tt.Args(MakeList("foo", "bar")).Rets(vs("foo", "bar"), nil),
		tt.Args(iterator{vs("a", "b")}).Rets(vs("a", "b"), nil),
		tt.Args(nonIterator{}).Rets(vs(), cannotIterate{"!!vals.nonIterator"}),
	})
}

// Iterate is tested indirectly by the test against Collect.

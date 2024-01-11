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
	tt.Test(t, CanIterate,
		Args("foo").Rets(true),
		Args(MakeList("foo", "bar")).Rets(true),
		Args(iterator{vs("a", "b")}).Rets(true),
		Args(nonIterator{}).Rets(false),
	)
}

func TestCollect(t *testing.T) {
	tt.Test(t, Collect,
		Args("foo").Rets(vs("f", "o", "o"), nil),
		Args(MakeList("foo", "bar")).Rets(vs("foo", "bar"), nil),
		Args(iterator{vs("a", "b")}).Rets(vs("a", "b"), nil),
		Args(nonIterator{}).Rets(vs(), cannotIterate{"!!vals.nonIterator"}),
	)
}

// Iterate is tested indirectly by the test against Collect.

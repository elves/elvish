package vals

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
	"github.com/elves/elvish/pkg/util"
)

// An implementation of Iterator.
type iterator struct{ elements []interface{} }

func (i iterator) Iterate(f func(interface{}) bool) {
	util.Feed(f, i.elements...)
}

// A non-implementation of Iterator.
type nonIterator struct{}

func TestCanIterate(t *testing.T) {
	Test(t, Fn("CanIterate", CanIterate), Table{
		Args("foo").Rets(true),
		Args(MakeList("foo", "bar")).Rets(true),
		Args(iterator{vs("a", "b")}).Rets(true),
		Args(nonIterator{}).Rets(false),
	})
}

func TestCollect(t *testing.T) {
	Test(t, Fn("Collect", Collect), Table{
		Args("foo").Rets(vs("f", "o", "o"), nil),
		Args(MakeList("foo", "bar")).Rets(vs("foo", "bar"), nil),
		Args(iterator{vs("a", "b")}).Rets(vs("a", "b"), nil),
		Args(nonIterator{}).Rets(vs(), cannotIterate{"!!vals.nonIterator"}),
	})
}

// Iterate is tested indirectly by the test against Collect.

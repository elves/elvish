package vals

import (
	"errors"
	"testing"

	"github.com/elves/elvish/tt"
	"github.com/elves/elvish/util"
)

// An implementation of Iterator.
type iterator struct{ elements []interface{} }

func (i iterator) Iterate(f func(interface{}) bool) {
	util.Feed(f, i.elements...)
}

// A non-implementation of Iterator.
type nonIterator struct{}

func TestCanIterate(t *testing.T) {
	tt.Test(t, tt.Fn("CanIterate", CanIterate), tt.Table{
		Args("foo").Rets(true),
		Args(MakeList("foo", "bar")).Rets(true),
		Args(iterator{vs("a", "b")}).Rets(true),
		Args(nonIterator{}).Rets(false),
	})
}

func TestCollect(t *testing.T) {
	tt.Test(t, tt.Fn("Collect", Collect), tt.Table{
		Args("foo").Rets(vs("f", "o", "o"), nil),
		Args(MakeList("foo", "bar")).Rets(vs("foo", "bar"), nil),
		Args(iterator{vs("a", "b")}).Rets(vs("a", "b"), nil),
		Args(nonIterator{}).Rets(vs(),
			errors.New("!!vals.nonIterator cannot be iterated")),
	})
}

// Iterate is tested indirectly by the test against Iterate.

package vals

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

func vs(xs ...interface{}) []interface{} { return xs }

type keysIterator struct{ keys []interface{} }

func (k keysIterator) IterateKeys(f func(interface{}) bool) {
	Feed(f, k.keys...)
}

type nonKeysIterator struct{}

func TestIterateKeys(t *testing.T) {
	Test(t, Fn("collectKeys", collectKeys), Table{
		Args(MakeMap("k1", "v1", "k2", "v2")).Rets(vs("k1", "k2"), nil),
		Args(keysIterator{vs("lorem", "ipsum")}).Rets(vs("lorem", "ipsum")),
		Args(nonKeysIterator{}).Rets(
			Any, cannotIterateKeysOf{"!!vals.nonKeysIterator"}),
	})
}

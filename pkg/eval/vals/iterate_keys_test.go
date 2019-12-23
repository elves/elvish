package vals

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/tt"
	"github.com/elves/elvish/pkg/util"
)

// A variant of IterateKeys that is easier to test.
func iterateKeys(v interface{}) ([]interface{}, error) {
	var keys []interface{}
	err := IterateKeys(v, func(k interface{}) bool {
		keys = append(keys, k)
		return true
	})
	return keys, err
}

func vs(xs ...interface{}) []interface{} { return xs }

type keysIterator struct{ keys []interface{} }

func (k keysIterator) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, k.keys...)
}

type nonKeysIterator struct{}

func TestIterateKeys(t *testing.T) {
	tt.Test(t, tt.Fn("iterateKeys", iterateKeys), tt.Table{
		Args(MakeMap("k1", "v1", "k2", "v2")).Rets(vs("k1", "k2"), nil),
		Args(testStructMap{}).Rets(vs("name", "score-number")),
		Args(keysIterator{vs("lorem", "ipsum")}).Rets(vs("lorem", "ipsum")),
		Args(nonKeysIterator{}).Rets(any,
			errors.New("!!vals.nonKeysIterator cannot have its keys iterated")),
	})
}

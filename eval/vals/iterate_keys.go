package vals

import (
	"errors"

	"github.com/xiaq/persistent/hashmap"
)

// KeysIterator wraps the IterateKeys method.
type KeysIterator interface {
	// IterateKeys calls the passed function with each key within the receiver.
	// The iteration is aborted if the function returns false.
	IterateKeys(func(v interface{}) bool)
}

// IterateKeys iterates the keys of the supplied value, calling the supplied
// function for each key. The function can return false to break the iteration.
// It is implemented for the mapKeysIterable type and types satisfying the
// IterateKeyser interface. For these types, it always returns a nil error. For
// other types, it doesn't do anything and returns an error.
func IterateKeys(v interface{}, f func(interface{}) bool) error {
	switch v := v.(type) {
	case KeysIterator:
		v.IterateKeys(f)
	case mapKeysIterable:
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, _ := it.Elem()
			if !f(k) {
				break
			}
		}
	default:
		return errors.New(Kind(v) + " cannot have its keys iterated")
	}
	return nil
}

type mapKeysIterable interface {
	Iterator() hashmap.Iterator
}

// Feed calls the function with given values, breaking earlier if the function
// returns false.
func Feed(f func(interface{}) bool, values ...interface{}) {
	for _, value := range values {
		if !f(value) {
			break
		}
	}
}

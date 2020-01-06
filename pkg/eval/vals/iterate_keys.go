package vals

import (
	"reflect"
)

// KeysIterator wraps the IterateKeys method.
type KeysIterator interface {
	// IterateKeys calls the passed function with each key within the receiver.
	// The iteration is aborted if the function returns false.
	IterateKeys(func(v interface{}) bool)
}

type cannotIterateKeysOf struct{ kind string }

func (err cannotIterateKeysOf) Error() string {
	return "cannot iterate keys of " + err.kind
}

// IterateKeys iterates the keys of the supplied value, calling the supplied
// function for each key. The function can return false to break the iteration.
// It is implemented for the Map type, StructMap types, and types satisfying the
// IterateKeyser interface. For these types, it always returns a nil error. For
// other types, it doesn't do anything and returns an error.
func IterateKeys(v interface{}, f func(interface{}) bool) error {
	switch v := v.(type) {
	case Map:
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, _ := it.Elem()
			if !f(k) {
				break
			}
		}
	case StructMap:
		for _, k := range getStructMapInfo(reflect.TypeOf(v)).fieldNames {
			if !f(k) {
				break
			}
		}
	case KeysIterator:
		v.IterateKeys(f)
	default:
		return cannotIterateKeysOf{Kind(v)}
	}
	return nil
}

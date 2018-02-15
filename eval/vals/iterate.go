package vals

import (
	"errors"

	"github.com/xiaq/persistent/vector"
)

// Iterator wraps the Iterate method.
type Iterator interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v interface{}) bool)
}

// Iterate iterates the supplied value, and calls the supplied function in each
// of its elements. The function can return false to break the iteration. It is
// implemented for the builtin type string, and types satisfying the
// listIterable or Iterator interface. For these types, it always returns a nil
// error. For other types, it doesn't do anything and returns an error.
func Iterate(v interface{}, f func(interface{}) bool) error {
	switch v := v.(type) {
	case Iterator:
		v.Iterate(f)
	case string:
		for _, r := range v {
			b := f(string(r))
			if !b {
				break
			}
		}
	case listIterable:
		for it := v.Iterator(); it.HasElem(); it.Next() {
			if !f(it.Elem()) {
				break
			}
		}
	default:
		return errors.New(Kind(v) + " cannot be iterated")
	}
	return nil
}

type listIterable interface {
	Iterator() vector.Iterator
}

var _ listIterable = vector.Vector(nil)

// Collect collects all elements of an iterable value into a slice.
func Collect(it interface{}) ([]interface{}, error) {
	var vs []interface{}
	if len := Len(it); len >= 0 {
		vs = make([]interface{}, 0, len)
	}
	err := Iterate(it, func(v interface{}) bool {
		vs = append(vs, v)
		return true
	})
	return vs, err
}

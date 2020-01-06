package vals

// Iterator wraps the Iterate method.
type Iterator interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v interface{}) bool)
}

type cannotIterate struct{ kind string }

func (err cannotIterate) Error() string { return "cannot iterate " + err.kind }

// CanIterate returns whether the value can be iterated. If CanIterate(v) is
// true, calling Iterate(v, f) will not result in an error.
func CanIterate(v interface{}) bool {
	switch v.(type) {
	case Iterator, string, List:
		return true
	}
	return false
}

// Iterate iterates the supplied value, and calls the supplied function in each
// of its elements. The function can return false to break the iteration. It is
// implemented for the builtin type string, the List type, and types satisfying
// the Iterator interface. For these types, it always returns a nil error. For
// other types, it doesn't do anything and returns an error.
func Iterate(v interface{}, f func(interface{}) bool) error {
	switch v := v.(type) {
	case string:
		for _, r := range v {
			b := f(string(r))
			if !b {
				break
			}
		}
	case List:
		for it := v.Iterator(); it.HasElem(); it.Next() {
			if !f(it.Elem()) {
				break
			}
		}
	case Iterator:
		v.Iterate(f)
	default:
		return cannotIterate{Kind(v)}
	}
	return nil
}

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

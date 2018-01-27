package types

import "reflect"

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other interface{}) bool
}

// Equal returns whether two values are equal. It is implemented for the builtin
// types bool and string, and types implementing the Equaler interface. For
// other types, it uses reflect.DeepEqual to compare the two values.
func Equal(x, y interface{}) bool {
	switch x := x.(type) {
	case bool:
		return x == y
	case string:
		return x == y
	case Equaler:
		return x.Equal(y)
	}
	return reflect.DeepEqual(x, y)
}

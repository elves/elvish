package types

import "reflect"

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other interface{}) bool
}

func Equal(x, y interface{}) bool {
	if equaler, ok := x.(Equaler); ok {
		return equaler.Equal(y)
	}
	return reflect.DeepEqual(x, y)
}

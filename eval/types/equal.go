package types

import (
	"reflect"
)

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other interface{}) bool
}

// Equal returns whether two values are equal. It is implemented for the builtin
// types bool and string, and types satisfying the listEqualable or Equaler
// interface. For other types, it uses reflect.DeepEqual to compare the two
// values.
func Equal(x, y interface{}) bool {
	switch x := x.(type) {
	case bool:
		return x == y
	case string:
		return x == y
	case listEqualable:
		var (
			yy listEqualable
			ok bool
		)
		if yy, ok = y.(listEqualable); !ok {
			return false
		}
		if x.Len() != yy.Len() {
			return false
		}
		ix := x.Iterator()
		iy := yy.Iterator()
		for ix.HasElem() && iy.HasElem() {
			if !Equal(ix.Elem(), iy.Elem()) {
				return false
			}
			ix.Next()
			iy.Next()
		}
		return true
	case Equaler:
		return x.Equal(y)
	}
	return reflect.DeepEqual(x, y)
}

type listEqualable interface {
	Lener
	listIterable
}

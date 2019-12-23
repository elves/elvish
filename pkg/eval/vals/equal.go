package vals

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
// types bool and string, the File, List, Map types, StructMap types, and types
// satisfying the Equaler interface. For other types, it uses reflect.DeepEqual
// to compare the two values.
func Equal(x, y interface{}) bool {
	switch x := x.(type) {
	case nil:
		return x == y
	case bool:
		return x == y
	case float64:
		return x == y
	case string:
		return x == y
	case File:
		if yy, ok := y.(File); ok {
			return x.Fd() == yy.Fd()
		}
		return false
	case List:
		if yy, ok := y.(List); ok {
			return equalList(x, yy)
		}
		return false
	case Map:
		if yy, ok := y.(Map); ok {
			return equalMap(x, yy)
		}
		return false
	case StructMap:
		if yy, ok := y.(StructMap); ok {
			return equalStructMap(x, yy)
		}
		return false
	case Equaler:
		return x.Equal(y)
	default:
		return reflect.DeepEqual(x, y)
	}
}

func equalList(x, y List) bool {
	if x.Len() != y.Len() {
		return false
	}
	ix := x.Iterator()
	iy := y.Iterator()
	for ix.HasElem() && iy.HasElem() {
		if !Equal(ix.Elem(), iy.Elem()) {
			return false
		}
		ix.Next()
		iy.Next()
	}
	return true
}

func equalMap(x, y Map) bool {
	if x.Len() != y.Len() {
		return false
	}
	for it := x.Iterator(); it.HasElem(); it.Next() {
		k, vx := it.Elem()
		vy, ok := y.Index(k)
		if !ok || !Equal(vx, vy) {
			return false
		}
	}
	return true
}

func equalStructMap(x, y StructMap) bool {
	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}
	xv := reflect.ValueOf(x)
	yv := reflect.ValueOf(y)
	n := xv.NumField()
	for i := 0; i < n; i++ {
		if !Equal(xv.Field(i).Interface(), yv.Field(i).Interface()) {
			return false
		}
	}
	return true
}

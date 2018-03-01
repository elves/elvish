package vals

import (
	"reflect"

	"github.com/xiaq/persistent/hashmap"
)

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other interface{}) bool
}

// Equal returns whether two values are equal. It is implemented for the builtin
// types bool and string, and types satisfying the listEqualable, mapEqualable
// or Equaler interface. For other types, it uses reflect.DeepEqual to compare
// the two values.
func Equal(x, y interface{}) bool {
	switch x := x.(type) {
	case Equaler:
		return x.Equal(y)
	case bool:
		return x == y
	case string:
		return x == y
	case listEqualable:
		if yy, ok := y.(listEqualable); ok {
			return equalList(x, yy)
		}
		return false
	case mapEqualable:
		if yy, ok := y.(mapEqualable); ok {
			return equalMap(x, yy)
		}
		return false
	}
	return reflect.DeepEqual(x, y)
}

type listEqualable interface {
	Lener
	listIterable
}

func equalList(x, y listEqualable) bool {
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

type mapEqualable interface {
	Lener
	Index(interface{}) (interface{}, bool)
	Iterator() hashmap.Iterator
}

var _ mapEqualable = hashmap.Map(nil)

func equalMap(x, y mapEqualable) bool {
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

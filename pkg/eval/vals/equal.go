package vals

import (
	"math/big"
	"reflect"

	"src.elv.sh/pkg/persistent/hashmap"
)

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other any) bool
}

// Equal returns whether two values are equal. It is implemented for the builtin
// types bool and string, the File, List, Map types, field map types, and types
// satisfying the Equaler interface. For other types, it uses reflect.DeepEqual
// to compare the two values.
func Equal(x, y any) bool {
	switch x := x.(type) {
	case nil:
		return x == y
	case bool:
		return x == y
	case int:
		return x == y
	case *big.Int:
		if y, ok := y.(*big.Int); ok {
			return x.Cmp(y) == 0
		}
		return false
	case *big.Rat:
		if y, ok := y.(*big.Rat); ok {
			return x.Cmp(y) == 0
		}
		return false
	case float64:
		return x == y
	case string:
		return x == y
	case List:
		if yy, ok := y.(List); ok {
			return equalList(x, yy)
		}
		return false
	// Types above are also handled in [Cmp]; keep the branches in the same
	// order.
	case Equaler:
		return x.Equal(y)
	case File:
		if yy, ok := y.(File); ok {
			return x.Fd() == yy.Fd()
		}
		return false
	case Map:
		switch y := y.(type) {
		case Map:
			return equalMap(x, y, Map.Iterator, Map.Index)
		default:
			if xKeys := GetFieldMapKeys(y); xKeys != nil {
				return equalFieldMapAndMap(y, xKeys, x)
			}
		}
		return false
	default:
		if xKeys := GetFieldMapKeys(x); xKeys != nil {
			switch y := y.(type) {
			case Map:
				return equalFieldMapAndMap(x, xKeys, y)
			default:
				if yKeys := GetFieldMapKeys(y); yKeys != nil {
					return equalFieldMapAndFieldMap(x, xKeys, y, yKeys)
				}
			}
		}
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

func equalMap[X, Y any, I hashmap.Iterator](x X, y Y, xit func(X) I, yidx func(Y, any) (any, bool)) bool {
	if Len(x) != Len(y) {
		return false
	}
	for it := xit(x); it.HasElem(); it.Next() {
		k, vx := it.Elem()
		vy, ok := yidx(y, k)
		if !ok || !Equal(vx, vy) {
			return false
		}
	}
	return true
}

func equalFieldMapAndMap(x any, xKeys FieldMapKeys, y Map) bool {
	if len(xKeys) != y.Len() {
		return false
	}
	xValue := reflect.ValueOf(x)
	for i, key := range xKeys {
		yValue, ok := y.Index(key)
		if !ok || !Equal(xValue.Field(i).Interface(), yValue) {
			return false
		}
	}
	return true
}

func equalFieldMapAndFieldMap(x any, xKeys FieldMapKeys, y any, yKeys FieldMapKeys) bool {
	if len(xKeys) != len(yKeys) {
		return false
	}
	xValue := reflect.ValueOf(x)
	for i, key := range xKeys {
		yValue, ok := indexFieldMap(y, key, yKeys)
		if !ok || !Equal(xValue.Field(i).Interface(), yValue) {
			return false
		}
	}
	return true
}

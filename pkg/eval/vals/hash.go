package vals

import (
	"math"
	"reflect"

	"github.com/xiaq/persistent/hash"
)

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

// Hash returns the 32-bit hash of a value. It is implemented for the builtin
// types bool and string, the File, List, Map types, StructMap types, and types
// satisfying the Hasher interface. For other values, it returns 0 (which is OK
// in terms of correctness).
func Hash(v interface{}) uint32 {
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case float64:
		return hash.UInt64(math.Float64bits(v))
	case string:
		return hash.String(v)
	case Hasher:
		return v.Hash()
	case File:
		return hash.UIntPtr(v.Fd())
	case List:
		h := hash.DJBInit
		for it := v.Iterator(); it.HasElem(); it.Next() {
			h = hash.DJBCombine(h, Hash(it.Elem()))
		}
		return h
	case Map:
		h := hash.DJBInit
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			h = hash.DJBCombine(h, Hash(k))
			h = hash.DJBCombine(h, Hash(v))
		}
		return h
	case StructMap:
		h := hash.DJBInit
		it := iterateStructMap(reflect.TypeOf(v))
		vValue := reflect.ValueOf(v)
		for it.Next() {
			_, field := it.Get(vValue)
			h = hash.DJBCombine(h, Hash(field))
		}
		return h
	}
	return 0
}

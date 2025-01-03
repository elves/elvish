package vals

import (
	"math"
	"math/big"
	"reflect"

	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/persistent/hashmap"
)

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

// Hash returns the 32-bit hash of a value. It is implemented for the builtin
// types bool and string, the File, List, Map types, field map types, and types
// satisfying the Hasher interface. For other values, it returns 0 (which is OK
// in terms of correctness).
func Hash(v any) uint32 {
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return hash.UIntPtr(uintptr(v))
	case *big.Int:
		h := hash.DJBCombine(hash.DJBInit, uint32(v.Sign()))
		for _, word := range v.Bits() {
			h = hash.DJBCombine(h, hash.UIntPtr(uintptr(word)))
		}
		return h
	case *big.Rat:
		return hash.DJB(Hash(v.Num()), Hash(v.Denom()))
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
		return hashMap(v.Iterator())
	default:
		if keys := GetFieldMapKeys(v); keys != nil {
			return hashFieldMap(v, keys)
		}
		return 0
	}
}

func hashMap(it hashmap.Iterator) uint32 {
	// The iteration order of maps only depends on the hash of the keys. It is
	// almost deterministic, with only one exception: when two keys have the
	// same hash, they get produced in insertion order. As a result, it is
	// possible for two maps that should be considered equal to produce entries
	// in different orders.
	//
	// So instead of using hash.DJBCombine, combine the hash result from each
	// key-value pair by summing, so that the order doesn't matter.
	//
	// TODO: This may not have very good hashing properties.
	var h uint32
	for ; it.HasElem(); it.Next() {
		k, v := it.Elem()
		h += hash.DJB(Hash(k), Hash(v))
	}
	return h
}

func hashFieldMap(v any, keys FieldMapKeys) uint32 {
	value := reflect.ValueOf(v)
	var h uint32
	for i, key := range keys {
		h += hash.DJB(Hash(key), Hash(value.Field(i).Interface()))
	}
	return h
}

package vals

import (
	"github.com/xiaq/persistent/hash"
)

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

// Hash returns the 32-bit hash of a value. It is implemented for the builtin
// types bool and string, and types satisfying the listIterable, mapIterable or
// Hasher interface. For other values, it returns 0 (which is OK in terms of
// correctness).
func Hash(v interface{}) uint32 {
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		return hash.String(v)
	case listIterable:
		h := hash.DJBInit
		for it := v.Iterator(); it.HasElem(); it.Next() {
			h = hash.DJBCombine(h, Hash(it.Elem()))
		}
		return h
	case mapIterable:
		h := hash.DJBInit
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			h = hash.DJBCombine(h, Hash(k))
			h = hash.DJBCombine(h, Hash(v))
		}
		return h
	case Hasher:
		return v.Hash()
	}
	return 0
}

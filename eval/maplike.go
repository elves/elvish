package eval

import "github.com/xiaq/persistent/hash"

type MapLike interface {
	Lener
	IndexOneer
	Assocer
	HasKeyer
	IterateKeyer
	IteratePairer
}

func eqMapLike(lhs MapLike, a interface{}) bool {
	rhs, ok := a.(MapLike)
	if !ok {
		return false
	}
	if lhs.Len() != rhs.Len() {
		return false
	}
	eq := true
	lhs.IteratePair(func(k, v Value) bool {
		if !v.Equal(rhs.IndexOne(k)) {
			eq = false
			return false
		}
		return true
	})
	return eq
}

func hashMapLike(m MapLike) uint32 {
	h := hash.DJBInit
	m.IteratePair(func(k, v Value) bool {
		// h = hash.DJBCombine(h, k.Hash())
		// h = hash.DJBCombine(h, v.Hash())
		return true
	})
	return h
}

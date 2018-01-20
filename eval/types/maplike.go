package types

import (
	"github.com/xiaq/persistent/hash"
)

type MapLike interface {
	Lener
	IndexOneer
	Assocer
	HasKeyer
	IterateKeyer
	IteratePairer
}

type HasKeyer interface {
	HasKey(k Value) bool
}

func EqMapLike(lhs MapLike, a interface{}) bool {
	rhs, ok := a.(MapLike)
	if !ok {
		return false
	}
	if lhs.Len() != rhs.Len() {
		return false
	}
	eq := true
	lhs.IteratePair(func(k, v Value) bool {
		v2, err := rhs.IndexOne(k)
		if err != nil || !v.Equal(v2) {
			eq = false
			return false
		}
		return true
	})
	return eq
}

func HashMapLike(m MapLike) uint32 {
	h := hash.DJBInit
	m.IteratePair(func(k, v Value) bool {
		h = hash.DJBCombine(h, k.Hash())
		h = hash.DJBCombine(h, v.Hash())
		return true
	})
	return h
}

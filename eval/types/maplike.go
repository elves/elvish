package types

import (
	"github.com/xiaq/persistent/hash"
)

type MapLike interface {
	Lener
	Indexer
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
		v2, err := rhs.Index(k)
		if err != nil || !Equal(v, v2) {
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
		h = hash.DJBCombine(h, Hash(k))
		h = hash.DJBCombine(h, Hash(v))
		return true
	})
	return h
}

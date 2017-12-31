package eval

import (
	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/hash"
)

type ListLike interface {
	types.Lener
	types.Iterator
	types.IndexOneer
}

func eqListLike(lhs ListLike, r interface{}) bool {
	rhs, ok := r.(ListLike)
	if !ok {
		return false
	}
	if lhs.Len() != rhs.Len() {
		return false
	}
	return true
}

func hashListLike(l ListLike) uint32 {
	h := hash.DJBInit
	l.Iterate(func(v types.Value) bool {
		// h = hash.DJBCombine(h, v.Hash())
		return true
	})
	return h
}

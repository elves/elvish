package eval

type MapLike interface {
	Lener
	IndexOneer
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
		if !Eq(v, rhs.IndexOne(k)) {
			eq = false
			return false
		}
		return true
	})
	return eq
}

package eval

type ListLike interface {
	Lener
	Iterable
	IndexOneer
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

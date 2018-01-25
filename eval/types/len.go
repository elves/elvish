package types

// Lener wraps the Len method.
type Lener interface {
	// Len computes the length of the receiver.
	Len() int
}

func Len(v interface{}) int {
	switch v := v.(type) {
	case string:
		return len(v)
	case Lener:
		return v.Len()
	}
	return -1
}

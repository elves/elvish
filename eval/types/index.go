package types

import "errors"

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(idx Value) (Value, error)
}

var errNotIndexable = errors.New("not indexable")

func Index(a, k Value) (Value, error) {
	switch a := a.(type) {
	case Indexer:
		return a.Index(k)
	default:
		return nil, errNotIndexable
	}
}

// MustIndex indexes i with k and returns the value. If the operation
// resulted in an error, it panics. It is useful when the caller knows that the
// key must be present.
func MustIndex(i Indexer, k Value) Value {
	v, err := i.Index(k)
	if err != nil {
		panic(err)
	}
	return v
}

package vals

import (
	"errors"
)

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves the value corresponding to the specified key in the
	// container. It returns the value (if any), and whether it actually exists.
	Index(k interface{}) (v interface{}, ok bool)
}

// ErrIndexer wraps the Index method.
type ErrIndexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(k interface{}) (interface{}, error)
}

var errNotIndexable = errors.New("not indexable")

type noSuchKeyError struct {
	key interface{}
}

// NoSuchKey returns an error indicating that a key is not found in a map-like
// value.
func NoSuchKey(k interface{}) error {
	return noSuchKeyError{k}
}

func (err noSuchKeyError) Error() string {
	return "no such key: " + Repr(err.key, NoPretty)
}

// Index indexes a value with the given key. It is implemented for types
// satisfying the ErrIndexer or Indexer interface, types satisfying the
// listIndexable interface (which covers Vector), and the builtin string type.
// For other types, it returns a nil value and a non-nil error.
func Index(a, k interface{}) (interface{}, error) {
	switch a := a.(type) {
	case ErrIndexer:
		return a.Index(k)
	case Indexer:
		v, ok := a.Index(k)
		if !ok {
			return nil, NoSuchKey(k)
		}
		return v, nil
	case listIndexable:
		return indexList(a, k)
	case string:
		return indexString(a, k)
	default:
		return nil, errNotIndexable
	}
}

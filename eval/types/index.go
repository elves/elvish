package types

import (
	"errors"
	"unicode/utf8"
)

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(idx Value) (Value, error)
}

var errNotIndexable = errors.New("not indexable")

func Index(a, k Value) (Value, error) {
	switch a := a.(type) {
	case string:
		i, j, err := indexString(a, k)
		if err != nil {
			return nil, err
		}
		return a[i:j], nil
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

func indexString(s string, idx Value) (int, int, error) {
	slice, i, j, err := ParseAndFixListIndex(ToString(idx), len(s))
	if err != nil {
		return 0, 0, err
	}
	r, size := utf8.DecodeRuneInString(s[i:])
	if r == utf8.RuneError {
		return 0, 0, ErrBadIndex
	}
	if slice {
		if r, _ := utf8.DecodeLastRuneInString(s[:j]); r == utf8.RuneError {
			return 0, 0, ErrBadIndex
		}
		return i, j, nil
	}
	return i, i + size, nil
}

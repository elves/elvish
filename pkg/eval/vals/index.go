package vals

import (
	"errors"
	"os"
)

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves the value corresponding to the specified key in the
	// container. It returns the value (if any), and whether it actually exists.
	Index(k any) (v any, ok bool)
}

// ErrIndexer wraps the Index method.
type ErrIndexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(k any) (any, error)
}

var errNotIndexable = errors.New("not indexable")

type noSuchKeyError struct {
	key any
}

// NoSuchKey returns an error indicating that a key is not found in a map-like
// value.
func NoSuchKey(k any) error {
	return noSuchKeyError{k}
}

func (err noSuchKeyError) Error() string {
	return "no such key: " + ReprPlain(err.key)
}

// Index indexes a value with the given key. It is implemented for the builtin
// type string, *os.File, List, StructMap and PseudoStructMap types, and types
// satisfying the ErrIndexer or Indexer interface (the Map type satisfies
// Indexer). For other types, it returns a nil value and a non-nil error.
func Index(a, k any) (any, error) {
	convertResult := func(v any, ok bool) (any, error) {
		if !ok {
			return nil, NoSuchKey(k)
		}
		return v, nil
	}
	switch a := a.(type) {
	case string:
		return indexString(a, k)
	case *os.File:
		return indexFile(a, k)
	case ErrIndexer:
		return a.Index(k)
	case Indexer:
		return convertResult(a.Index(k))
	case List:
		return indexList(a, k)
	case StructMap:
		return convertResult(indexStructMap(a, k))
	case PseudoMap:
		return convertResult(indexStructMap(a.Fields(), k))
	default:
		return nil, errNotIndexable
	}
}

func indexFile(f *os.File, k any) (any, error) {
	switch k {
	case "fd":
		return int(f.Fd()), nil
	case "name":
		return f.Name(), nil
	}
	return nil, NoSuchKey(k)
}

func indexStructMap(a StructMap, k any) (any, bool) {
	fieldName, ok := k.(string)
	if !ok || fieldName == "" {
		return nil, false
	}
	for it := iterateStructMap(a); it.HasElem(); it.Next() {
		k, v := it.elem()
		if k == fieldName {
			return FromGo(v), true
		}
	}
	return nil, false
}

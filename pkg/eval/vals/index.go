package vals

import (
	"errors"
	"os"
	"reflect"
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

// TODO: Replace this with a a generalized introspection mechanism based on PseudoStructMap for
// *os.File objects so that commands like `keys` also work on those objects.
var errInvalidOsFileIndex = errors.New("invalid index for a File object")

func indexOsFile(f *os.File, k interface{}) (interface{}, error) {
	switch k := k.(type) {
	case string:
		switch {
		case k == "fd":
			return int(f.Fd()), nil
		case k == "name":
			return f.Name(), nil
		default:
			return nil, errInvalidOsFileIndex
		}
	default:
		return nil, errInvalidOsFileIndex
	}
}

// Index indexes a value with the given key. It is implemented for the builtin
// type string, the List type, StructMap types, and types satisfying the
// ErrIndexer or Indexer interface (the Map type satisfies Indexer). For other
// types, it returns a nil value and a non-nil error.
func Index(a, k interface{}) (interface{}, error) {
	switch a := a.(type) {
	case string:
		return indexString(a, k)
	case *os.File:
		return indexOsFile(a, k)
	case ErrIndexer:
		return a.Index(k)
	case Indexer:
		v, ok := a.Index(k)
		if !ok {
			return nil, NoSuchKey(k)
		}
		return v, nil
	case List:
		return indexList(a, k)
	case StructMap:
		return indexStructMap(a, k)
	case PseudoStructMap:
		return indexStructMap(a.Fields(), k)
	default:
		return nil, errNotIndexable
	}
}

func indexStructMap(a StructMap, k interface{}) (interface{}, error) {
	fieldName, ok := k.(string)
	if !ok || fieldName == "" {
		return nil, NoSuchKey(k)
	}
	aValue := reflect.ValueOf(a)
	it := iterateStructMap(reflect.TypeOf(a))
	for it.Next() {
		k, v := it.Get(aValue)
		if k == fieldName {
			return FromGo(v), nil
		}
	}
	return nil, NoSuchKey(fieldName)
}

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
// type string, *os.File, List, field map and pseudo map types, and types
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
	case PseudoMap:
		return convertResult(indexMethodMap(a.Fields(), k))
	default:
		if keys := GetFieldMapKeys(a); keys != nil {
			return convertResult(indexFieldMap(a, k, keys))
		} else {
			return nil, errNotIndexable
		}
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

func indexMethodMap(m MethodMap, k any) (any, bool) {
	kstring, ok := k.(string)
	if !ok {
		return nil, false
	}
	for i, key := range getMethodMapKeys(m) {
		if kstring == key {
			return reflect.ValueOf(m).Method(i).Call(nil)[0].Interface(), true
		}
	}
	return nil, false
}

func indexFieldMap(m, k any, keys FieldMapKeys) (any, bool) {
	kstring, ok := k.(string)
	if !ok {
		return nil, false
	}
	for i, key := range keys {
		if kstring == key {
			return reflect.ValueOf(m).Field(i).Interface(), true
		}
	}
	return nil, false
}

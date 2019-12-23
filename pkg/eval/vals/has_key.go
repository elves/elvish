package vals

import (
	"reflect"

	"github.com/xiaq/persistent/hashmap"
)

// HasKeyer wraps the HasKey method.
type HasKeyer interface {
	// HasKey returns whether the receiver has the given argument as a valid
	// key.
	HasKey(interface{}) bool
}

// HasKey returns whether a container has a key. It is implemented for the Map
// type, StructMap types, and types satisfying the HasKeyer interface. It falls
// back to iterating keys using IterateKeys, and if that fails, it falls back to
// calling Len and checking if key is a valid numeric or slice index. Otherwise
// it returns false.
func HasKey(container, key interface{}) bool {
	switch container := container.(type) {
	case Map:
		return hashmap.HasKey(container, key)
	case StructMap:
		kstring, ok := key.(string)
		if !ok {
			return false
		}
		for _, fieldName := range getStructMapInfo(reflect.TypeOf(container)).fieldNames {
			if fieldName == kstring {
				return true
			}
		}
		return false
	case HasKeyer:
		return container.HasKey(key)
	default:
		var found bool
		err := IterateKeys(container, func(k interface{}) bool {
			if key == k {
				found = true
			}
			return !found
		})
		if err == nil {
			return found
		}
		if len := Len(container); len >= 0 {
			// XXX(xiaq): Not all types that implement Lener have numerical indices
			_, err := ConvertListIndex(key, len)
			return err == nil
		}
		return false
	}
}

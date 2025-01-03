package vals

import (
	"src.elv.sh/pkg/persistent/hashmap"
)

// HasKeyer wraps the HasKey method.
type HasKeyer interface {
	// HasKey returns whether the receiver has the given argument as a valid
	// key.
	HasKey(any) bool
}

// HasKey returns whether a container has a key. It is implemented for the Map
// type, field map types, and types satisfying the HasKeyer interface. It falls
// back to iterating keys using IterateKeys, and if that fails, it falls back to
// calling Len and checking if key is a valid numeric or slice index. Otherwise
// it returns false.
func HasKey(container, key any) bool {
	switch container := container.(type) {
	case HasKeyer:
		return container.HasKey(key)
	case Map:
		return hashmap.HasKey(container, key)
	case PseudoMap:
		return hasKeyFieldOrMethodMap(key, getMethodMapKeys(container.Fields()))
	default:
		if keys := GetFieldMapKeys(container); keys != nil {
			return hasKeyFieldOrMethodMap(key, keys)
		} else {
			return hasKeyViaIterateKeys(container, key)
		}
	}
}

func hasKeyFieldOrMethodMap(k any, keys []string) bool {
	kstring, ok := k.(string)
	if !ok || kstring == "" {
		return false
	}
	for _, fieldName := range keys {
		if kstring == fieldName {
			return true
		}
	}
	return false
}

func hasKeyViaIterateKeys(container, key any) bool {
	var found bool
	err := IterateKeys(container, func(k any) bool {
		if key == k {
			found = true
		}
		return !found
	})
	if err == nil {
		return found
	}
	if len := Len(container); len >= 0 {
		// TODO(xiaq): Not all types that implement Lener have numerical
		// indices
		_, err := ConvertListIndex(key, len)
		return err == nil
	}
	return false
}

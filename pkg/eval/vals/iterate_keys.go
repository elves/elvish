package vals

// KeysIterator wraps the IterateKeys method.
type KeysIterator interface {
	// IterateKeys calls the passed function with each key within the receiver.
	// The iteration is aborted if the function returns false.
	IterateKeys(func(v any) bool)
}

type cannotIterateKeysOf struct{ kind string }

func (err cannotIterateKeysOf) Error() string {
	return "cannot iterate keys of " + err.kind
}

// IterateKeys iterates the keys of the supplied value, calling the supplied
// function for each key. The function can return false to break the iteration.
// It is implemented for the Map type, field map types, and types satisfying the
// IterateKeyser interface. For these types, it always returns a nil error. For
// other types, it doesn't do anything and returns an error.
func IterateKeys(v any, f func(any) bool) error {
	switch v := v.(type) {
	case KeysIterator:
		v.IterateKeys(f)
	case Map:
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, _ := it.Elem()
			if !f(k) {
				break
			}
		}
	case PseudoMap:
		iterateKeysFieldOrMethodMap(getMethodMapKeys(v.Fields()), f)
	default:
		if keys := GetFieldMapKeys(v); keys != nil {
			iterateKeysFieldOrMethodMap(keys, f)
		} else {
			return cannotIterateKeysOf{Kind(v)}
		}
	}
	return nil
}

func iterateKeysFieldOrMethodMap(keys []string, f func(any) bool) {
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}

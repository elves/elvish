package vals

import (
	"reflect"
	"sync"

	"src.elv.sh/pkg/strutil"
)

// IsFieldMap returns whether a value is a field map.
//
// A field map is a struct with at least one field, and all fields are exported
// and non-anonymous. In Elvish code, it behaves exactly the same as a map: the
// keys are dash-case versions of the field names, and the values are the field
// value.
func IsFieldMap(v any) bool {
	return GetFieldMapKeys(v) != nil
}

func promoteFieldMapToMap(v any, keys FieldMapKeys) Map {
	m := EmptyMap
	value := reflect.ValueOf(v)
	for i, key := range keys {
		m = m.Assoc(key, value.Field(i).Interface())
	}
	return m
}

var fieldMapKeysCache sync.Map

// FieldMapKeys contains the map keys corresponding to each field in a field
// map. A nil value means that a type is not a field map.
type FieldMapKeys []string

// GetFieldMapKeys returns the FieldMapKeys for v if it's a field map, or nil if
// it's not.
func GetFieldMapKeys(v any) FieldMapKeys {
	return getFieldMapKeysT(TypeOf(v))
}

func getFieldMapKeysT(t reflect.Type) FieldMapKeys {
	if t.Kind() != reflect.Struct {
		return nil
	}
	if keys, ok := fieldMapKeysCache.Load(t); ok {
		return keys.(FieldMapKeys)
	}
	keys := makeFieldMapKeys(t)
	fieldMapKeysCache.Store(t, keys)
	return keys
}

func makeFieldMapKeys(t reflect.Type) FieldMapKeys {
	n := t.NumField()
	if n == 0 {
		return nil
	}
	keys := make([]string, n)
	for i := range n {
		field := t.Field(i)
		if field.PkgPath != "" || field.Anonymous {
			return nil
		}
		keys[i] = strutil.CamelToDashed(field.Name)
	}
	return keys
}

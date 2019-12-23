package vals

import "reflect"

// StructMap wraps the IsStructMap method, a marker to make some Elvish
// operations work on structs via reflection. The operations are Kind, Equal,
// Hash, Repr, Len, Index, Assoc, IterateKeys and HasKey.
//
// The marker method must be implemented on the value type of a struct (not the
// pointer, or any other type), otherwise operations on such values may panic.
// The struct may not contain cyclic data, and all the fields of the struct must
// be exported. Elvish reuses the `json:"name" annotations of the fields as the
// name of the Elvish fields in the Elvish operations.
type StructMap interface {
	IsStructMap(StructMapMarker)
}

// StructMapMarker is used in the marker method IsStructMap.
type StructMapMarker struct{}

// Keeps cached information about a StructMap.
type structMapInfo struct {
	fieldNames []string
}

var structMapInfos = make(map[reflect.Type]structMapInfo)

// Gets the structMapInfo associated with a type, caching the result.
func getStructMapInfo(t reflect.Type) structMapInfo {
	if info, ok := structMapInfos[t]; ok {
		return info
	}
	info := makeStructMapInfo(t)
	structMapInfos[t] = info
	return info
}

func makeStructMapInfo(t reflect.Type) structMapInfo {
	n := t.NumField()
	fieldNames := make([]string, n)
	for i := 0; i < n; i++ {
		fieldNames[i] = t.Field(i).Tag.Get("json")
	}
	return structMapInfo{fieldNames}
}

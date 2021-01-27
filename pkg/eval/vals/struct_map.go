package vals

import (
	"reflect"
	"sync"

	"src.elv.sh/pkg/strutil"
)

// StructMap may be implemented by a struct to mark the struct as a "struct
// map", which causes Elvish to treat it like a read-only map. Each exported,
// named field and getter method (a method taking no argument and returning one
// value) becomes a field of the map, with the name mapped to dash-case.
//
// The following operations are derived for structmaps: Kind, Repr, Hash, Len,
// Index, HasKey and IterateKeys.
//
// Example:
//
//   type someStruct struct {
//       FooBar int
//       lorem  string
//   }
//
//   func (someStruct) IsStructMap() { }
//
//   func (s SomeStruct) Ipsum() string { return s.lorem }
//
//   func (s SomeStruct) OtherMethod(int) { }
//
// An instance of someStruct behaves like a read-only map with 3 fields:
// foo-bar, lorem and ipsum.
type StructMap interface{ IsStructMap() }

// PseudoStructMap may be implemented by a type to derive the Repr, Index,
// HasKey and IterateKeys operations from the struct map returned by the Fields
// method.
type PseudoStructMap interface{ Fields() StructMap }

// Keeps cached information about a structMap.
type structMapInfo struct {
	filledFields int
	plainFields  int
	// Dash-case names for all fields. The first plainFields elements
	// corresponds to all the plain fields, while the rest corresponds to getter
	// fields. May contain empty strings if the corresponding field is not
	// reflected onto the structMap (i.e. unexported fields, unexported methods
	// and non-getter methods).
	fieldNames []string
}

var structMapInfos sync.Map

// Gets the structMapInfo associated with a type, caching the result.
func getStructMapInfo(t reflect.Type) structMapInfo {
	if info, ok := structMapInfos.Load(t); ok {
		return info.(structMapInfo)
	}
	info := makeStructMapInfo(t)
	structMapInfos.Store(t, info)
	return info
}

func makeStructMapInfo(t reflect.Type) structMapInfo {
	n := t.NumField()
	m := t.NumMethod()
	fieldNames := make([]string, n+m)
	filledFields := 0

	for i := 0; i < n; i++ {
		field := t.Field(i)
		if field.PkgPath == "" && !field.Anonymous {
			fieldNames[i] = strutil.CamelToDashed(field.Name)
			filledFields++
		}
	}

	for i := 0; i < m; i++ {
		method := t.Method(i)
		if method.PkgPath == "" && method.Type.NumIn() == 1 && method.Type.NumOut() == 1 {
			fieldNames[i+n] = strutil.CamelToDashed(method.Name)
			filledFields++
		}
	}

	return structMapInfo{filledFields, n, fieldNames}
}

type structMapIterator struct {
	info structMapInfo
	i    int
}

func iterateStructMap(t reflect.Type) *structMapIterator {
	return &structMapIterator{getStructMapInfo(t), -1}
}

func (it *structMapIterator) Next() bool {
	fields := it.info.fieldNames
	if it.i >= len(fields) {
		return false
	}

	it.i++
	for it.i < len(fields) && fields[it.i] == "" {
		it.i++
	}
	return it.i < len(fields)
}

func (it *structMapIterator) Get(v reflect.Value) (string, interface{}) {
	name := it.info.fieldNames[it.i]
	if it.i < it.info.plainFields {
		return name, v.Field(it.i).Interface()
	}
	method := v.Method(it.i - it.info.plainFields)
	return name, method.Call(nil)[0].Interface()
}

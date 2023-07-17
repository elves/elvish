package vals

import (
	"reflect"
	"sync"

	"src.elv.sh/pkg/strutil"
)

// StructMap may be implemented by a struct to make it accessible to Elvish code
// as a map. Each exported, named field and getter method (a method taking no
// argument and returning one value) becomes a field of the map, with the name
// mapped to dash-case.
//
// Struct maps are indistinguishable from normal maps for Elvish code. The
// operations Kind, Repr, Hash, Equal, Len, Index, HasKey and IterateKeys handle
// struct maps consistently with maps; the Assoc and Dissoc operations convert
// struct maps to maps.
//
// Example:
//
//	type someStruct struct {
//	    // Provides the "foo-bar" field
//	    FooBar int
//	    lorem  string
//	}
//
//	// Marks someStruct as a struct map
//	func (someStruct) IsStructMap() { }
//
//	// Provides the "ipsum" field
//	func (s SomeStruct) Ipsum() string { return s.lorem }
//
//	// Not a getter method; doesn't provide any field
//	func (s SomeStruct) OtherMethod(int) { }
type StructMap interface{ IsStructMap() }

func promoteToMap(v StructMap) Map {
	m := EmptyMap
	for it := iterateStructMap(v); it.HasElem(); it.Next() {
		m = m.Assoc(it.Elem())
	}
	return m
}

// PseudoMap may be implemented by a type to support map-like introspection. The
// Repr, Index, HasKey and IterateKeys operations handle pseudo maps.
type PseudoMap interface{ Fields() StructMap }

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
	m     reflect.Value
	info  structMapInfo
	index int
}

func iterateStructMap(m StructMap) *structMapIterator {
	it := &structMapIterator{reflect.ValueOf(m), getStructMapInfo(reflect.TypeOf(m)), 0}
	it.fixIndex()
	return it
}

func (it *structMapIterator) fixIndex() {
	fieldNames := it.info.fieldNames
	for it.index < len(fieldNames) && fieldNames[it.index] == "" {
		it.index++
	}
}

func (it *structMapIterator) Elem() (any, any) {
	return it.elem()
}

func (it *structMapIterator) elem() (string, any) {
	name := it.info.fieldNames[it.index]
	if it.index < it.info.plainFields {
		return name, it.m.Field(it.index).Interface()
	}
	method := it.m.Method(it.index - it.info.plainFields)
	return name, method.Call(nil)[0].Interface()
}

func (it *structMapIterator) HasElem() bool {
	return it.index < len(it.info.fieldNames)
}

func (it *structMapIterator) Next() {
	it.index++
	it.fixIndex()
}

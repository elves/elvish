package vals

import (
	"reflect"

	"src.elv.sh/pkg/strutil"
)

// Note: This file is a stub for now.
//
// Actually implementing this requires calling GoFn,
// introducing a circular dependency between this package and the eval package.

// Maps method index to its Elvish map key.
type MethodMapKeys []string

var methodMapKeysForType map[reflect.Type]MethodMapKeys

func GetMethodMapKeys(v any) MethodMapKeys {
	return methodMapKeysForType[TypeOf(v)]
}

// Registers a concrete (i.e. non-interface) type as a method map.
//
// A method map behaves exactly like a str->fn map in Elvish,
// with each method corresponding to a key-value pair:
//
//   - The key is the dash-case version of the method name.
//   - The value is the [GoFn] wrapping of the method implementation.
//
// All methods must be exported.
//
// It's comparable to the concept of field maps,
// but unlike field maps, method maps require explicit registration.
// This is necessary because too many types are eligible as a method map,
// but we don't want to expose all of them in this way.
//
// For the mechanism that enables bindings in the other direction -
// using an Elvish map as an interface implemention,
// see [etk.RegisterStructOfFuncForInterface].
//
// This functions returns a value so that it can be called from the top level:
//
//	var _ = vals.RegisterMethodMap[MyType]()
func RegisterMethodMap[T any]() struct{} {
	typ := reflect.TypeFor[T]()
	if typ.Kind() == reflect.Interface {
		panic("cannot register interface type as method map")
	}
	n := typ.NumMethod()
	keys := make([]string, n)
	for i := range n {
		method := typ.Method(i)
		if method.PkgPath != "" {
			panic("All methods of a method map must be exported; found unexported method " + method.Name)
		}
		keys[i] = strutil.CamelToDashed(method.Name)
	}
	methodMapKeysForType[typ] = keys
	return struct{}{}
}

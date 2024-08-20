package vals

import (
	"fmt"
	"reflect"
	"sync"

	"src.elv.sh/pkg/strutil"
)

// PseudoMap may be implemented by a type to support map-like introspection. The
// Repr, Index, HasKey and IterateKeys operations handle pseudo maps.
type PseudoMap interface{ Fields() MethodMap }

// MethodMap is a type whose methods are all nullary and exported.
type MethodMap any

var methodMapKeysCache sync.Map

type methodMapKeys []string

func getMethodMapKeys(v MethodMap) methodMapKeys {
	t := reflect.TypeOf(v)
	if fields, ok := methodMapKeysCache.Load(t); ok {
		return fields.(methodMapKeys)
	}
	keys := makeMethodMapKeys(t)
	methodMapKeysCache.Store(t, keys)
	return keys
}

func makeMethodMapKeys(t reflect.Type) methodMapKeys {
	n := t.NumMethod()
	keys := make([]string, n)
	for i := range n {
		method := t.Method(i)
		if method.PkgPath != "" || method.Type.NumIn() != 1 || method.Type.NumOut() != 1 {
			// Unlike [getFieldMapKeys] and [makeFieldMapKeys], this function is
			// only called when a type implements the [PseudoMap] interface and
			// thus references [MethodMap] explicitly, so we can assume that any
			// mismatch in expectation is a programmer error.
			panic(fmt.Sprintf("method %v is not exported nullary", method))
		}
		keys[i] = strutil.CamelToDashed(method.Name)
	}
	return keys
}

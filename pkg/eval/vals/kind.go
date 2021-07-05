package vals

import (
	"fmt"
	"math/big"
	"reflect"
)

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Kind returns the "kind" of the value, a concept similar to type but not yet very well defined. It
// is implemented for the builtin nil, bool string, File, List, Map, StructMap, eval.Callable types,
// and types satisfying the Kinder interface. For other types, it returns the Go type name of the
// argument prefaced with "!!".
func Kind(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return "nil"
	case bool, *bool:
		return "bool"
	case string, *string:
		return "string"
	case int, *int, *big.Int, *big.Rat, float64, *float64:
		return "number"
	case Kinder:
		// We want to call Kind() if it is available even if the object might otherwise match one of
		// the other Elvish types below. This isn't first in the list because placing it there hurts
		// performance but is otherwise benign.
		return v.Kind()
	case File, *File:
		return "file"
	case List, *List:
		return "list"
	case Map, *Map:
		return "map"
	case StructMap, *StructMap, PseudoStructMap, *PseudoStructMap:
		return "structmap"
	default:
		// Try to dynamically invoke the Kind() method if it exists for the object. One such type is
		// eval.Callable. We can't include a "case eval.Callable:" above since that would create an
		// import dependency loop. The Kinder case above doesn't handle that type, even though it
		// implements the Kinder interface, because eval.Callable is an interface rather than
		// concrete type when used in callback APIs like vars.PtrVar.
		var m reflect.Value
		switch reflect.TypeOf(v).Kind() {
		case reflect.Ptr, reflect.Interface:
			m = reflect.ValueOf(v).Elem().MethodByName("Kind")
		default:
			m = reflect.ValueOf(v).MethodByName("Kind")
		}
		if m.IsValid() {
			retvals := m.Call(nil)
			return retvals[0].String()
		}
		// Fallback to the special "!!" syntax to make it clear something is wrong since we were
		// unable to provide a meaningful "kind" for the object and had to resort to returning the
		// symbolic Go type.
		return fmt.Sprintf("!!%T", v)
	}
}

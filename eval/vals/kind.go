package vals

import (
	"fmt"
)

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Kind returns the "kind" of the value, a concept similar to type but not yet
// very well defined. It is implemented for the builtin nil, bool and string,
// the File, List, Map types, StructMap types, and types satisfying the Kinder
// interface. For other types, it returns the Go type name of the argument
// preceded by "!!".
func Kind(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case string:
		return "string"
	case File:
		return "file"
	case List:
		return "list"
	case Map:
		return "map"
	case StructMap:
		return "structmap"
	case Kinder:
		return v.Kind()
	default:
		return fmt.Sprintf("!!%T", v)
	}
}

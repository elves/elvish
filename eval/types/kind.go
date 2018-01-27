package types

import "fmt"

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Kind returns the "kind" of the value, a concept similar to type but not yet
// very well defined. It is implemented for the builtin types bool and string,
// and types implementing the Kinder interface. For other types, it returns the
// Go type name of the argument preceeded by "!!".
func Kind(v interface{}) string {
	switch v := v.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case Kinder:
		return v.Kind()
	default:
		return fmt.Sprintf("!!%T", v)
	}
}

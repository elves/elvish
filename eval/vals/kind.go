package vals

import (
	"fmt"

	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Kind returns the "kind" of the value, a concept similar to type but not yet
// very well defined. It is implemented for the builtin nil, bool and string,
// the Vector and Map types, and types implementing the Kinder interface. For
// other types, it returns the Go type name of the argument preceded by "!!".
func Kind(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case string:
		return "string"
	case vector.Vector:
		return "list"
	case hashmap.Map:
		return "map"
	case Kinder:
		return v.Kind()
	default:
		return fmt.Sprintf("!!%T", v)
	}
}

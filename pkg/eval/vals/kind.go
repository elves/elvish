package vals

import (
	"fmt"
	"math/big"
)

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Kind returns the "kind" of the value, a concept similar to type but not yet
// very well defined. It is implemented for the builtin nil, bool and string,
// the File, List, Map types, field map types, and types satisfying the Kinder
// interface. For other types, it returns the Go type name of the argument
// preceded by "!!".
//
// TODO: Decide what `kind-of` should report for an external command object
// and document the rationale for the choice in the doc string for `func
// (ExternalCmd) Kind()` as well as user facing documentation. It's not
// obvious why this returns "fn" rather than "external" for that case.
func Kind(v any) string {
	switch v := v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case string:
		return "string"
	case int, *big.Int, *big.Rat, float64:
		return "number"
	case File:
		return "file"
	case List:
		return "list"
	case Map:
		return "map"
	case Kinder:
		return v.Kind()
	default:
		if IsFieldMap(v) {
			return "map"
		}
		return fmt.Sprintf("!!%T", v)
	}
}

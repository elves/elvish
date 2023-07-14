package vals

import (
	"reflect"

	"src.elv.sh/pkg/persistent/vector"
)

// Lener wraps the Len method.
type Lener interface {
	// Len computes the length of the receiver.
	Len() int
}

var _ Lener = vector.Vector(nil)

// Len returns the length of the value, or -1 if the value does not have a
// well-defined length. It is implemented for the builtin type string, StructMap
// types, and types satisfying the Lener interface. For other types, it returns
// -1.
func Len(v any) int {
	switch v := v.(type) {
	case string:
		return len(v)
	case Lener:
		return v.Len()
	case StructMap:
		return lenStructMap(v)
	}
	return -1
}

func lenStructMap(m StructMap) int {
	return getStructMapInfo(reflect.TypeOf(m)).filledFields
}

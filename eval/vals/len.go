package vals

import (
	"reflect"

	"github.com/xiaq/persistent/vector"
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
func Len(v interface{}) int {
	switch v := v.(type) {
	case string:
		return len(v)
	case StructMap:
		return reflect.ValueOf(v).NumField()
	case Lener:
		return v.Len()
	}
	return -1
}

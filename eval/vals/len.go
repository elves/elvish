package vals

import "github.com/xiaq/persistent/vector"

// Lener wraps the Len method.
type Lener interface {
	// Len computes the length of the receiver.
	Len() int
}

var _ Lener = vector.Vector(nil)

// Len returns the length of the value, or -1 if the value does not have a
// well-defined length. It is implemented for the builtin string type and types
// satisfying the Lener interface. For other types, it returns -1.
func Len(v interface{}) int {
	switch v := v.(type) {
	case Lener:
		return v.Len()
	case string:
		return len(v)
	}
	return -1
}

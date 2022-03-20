package vals

// Booler wraps the Bool method.
type Booler interface {
	// Bool computes the truth value of the receiver.
	Bool() bool
}

// Bool converts a value to bool. It is implemented for nil, the builtin bool
// type, and types implementing the Booler interface. For all other values, it
// returns true.
func Bool(v any) bool {
	switch v := v.(type) {
	case nil:
		return false
	case bool:
		return v
	case Booler:
		return v.Bool()
	}
	return true
}

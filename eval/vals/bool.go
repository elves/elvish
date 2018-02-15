package vals

// Booler wraps the Bool method.
type Booler interface {
	// Bool computes the truth value of the receiver.
	Bool() bool
}

// Bool converts a value to bool. It is implemented for the builtin bool type
// and types implementing the Booler interface. For all other values, it returns
// true.
func Bool(v interface{}) bool {
	switch v := v.(type) {
	case Booler:
		return v.Bool()
	case bool:
		return v
	}
	return true
}

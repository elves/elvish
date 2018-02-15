package vals

// Stringer wraps the String method.
type Stringer interface {
	// Stringer converts the receiver to a string.
	String() string
}

// ToString converts a Value to string. When the Value type implements
// String(), it is used. Otherwise Repr(NoPretty) is used.
func ToString(v interface{}) string {
	switch v := v.(type) {
	case Stringer:
		return v.String()
	case string:
		return v
	default:
		return Repr(v, NoPretty)
	}
}

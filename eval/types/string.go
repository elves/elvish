package types

// Stringer wraps the String method.
type Stringer interface {
	// Stringer converts the receiver to a string.
	String() string
}

// ToString converts a Value to string. When the Value type implements
// String(), it is used. Otherwise Repr(NoPretty) is used.
func ToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case Stringer:
		return v.String()
	default:
		return Repr(v, NoPretty)
	}
}

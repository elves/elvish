package vals

import "strconv"

// Stringer wraps the String method.
type Stringer interface {
	// Stringer converts the receiver to a string.
	String() string
}

// ToString converts a Value to string. It is implemented for the builtin
// float64 and string types, and type satisfying the Stringer interface. It
// falls back to Repr(v, NoPretty).
func ToString(v interface{}) string {
	switch v := v.(type) {
	case float64:
		return formatFloat64(v)
	case string:
		return v
	case Stringer:
		return v.String()
	default:
		return Repr(v, NoPretty)
	}
}

func formatFloat64(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

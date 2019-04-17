package vals

import "strconv"

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
	case float64:
		return formatFloat64(v)
	default:
		return Repr(v, NoPretty)
	}
}

func formatFloat64(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

package vals

import (
	"strconv"
	"strings"
)

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
	// Go's 'g' format is almost what we want, except that its threshold for
	// "large exponent" is too low - 6 for positive exponents and -5 for
	// negative exponents. This means that relative small numbers like 1234567
	// are printed with scientific notations, something we don't really want.
	// See also b.elv.sh/811.
	//
	// So we emulate the 'g' format by first using 'f', parse the result, and
	// use 'e' if the exponent >= 14 or <= -5.
	s := strconv.FormatFloat(f, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if (i == -1 && len(s) > 14) || i >= 14 || strings.HasPrefix(s, "0.0000") {
		return strconv.FormatFloat(f, 'e', -1, 64)
	}
	return s
}

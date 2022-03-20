package vals

import (
	"math"
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
// falls back to Repr(v).
func ToString(v any) string {
	switch v := v.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return formatFloat64(v)
		// Other number types handled by "case Stringer"
	case string:
		return v
	case Stringer:
		return v.String()
	default:
		return ReprPlain(v)
	}
}

func formatFloat64(f float64) string {
	// Go's 'g' format is not quite ideal for printing floating point numbers;
	// it uses scientific notation too aggressively, and relatively small
	// numbers like 1234567 are printed with scientific notations, something we
	// don't really want.
	//
	// So we use a different algorithm for determining when to use scientific
	// notation. The algorithm is reverse-engineered from Racket's; it may not
	// be a perfect clone but hopefully good enough.
	//
	// See also b.elv.sh/811 for more context.
	s := strconv.FormatFloat(f, 'f', -1, 64)
	noPoint := !strings.ContainsRune(s, '.')
	if (noPoint && len(s) > 14 && s[len(s)-1] == '0') ||
		strings.HasPrefix(s, "0.0000") {
		return strconv.FormatFloat(f, 'e', -1, 64)
	} else if noPoint && !math.IsNaN(f) && !math.IsInf(f, 0) {
		return s + ".0"
	}
	return s
}

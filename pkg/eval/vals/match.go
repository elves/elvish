package vals

import (
	"regexp"
)

// MatchesRegexp returns whether The first value matches the second value interpreted as a regexp.
// Both bytes must be a string.
func MatchesRegexp(x, y interface{}) bool {
	val, ok := x.(string)
	if !ok {
		return false
	}
	pat, ok := y.(string)
	if !ok {
		return false
	}

	matched, err := regexp.MatchString(pat, val)
	if err != nil {
		return false
	}
	return matched
}

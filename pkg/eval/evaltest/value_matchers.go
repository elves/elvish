package evaltest

import (
	"math"
	"regexp"
)

// ValueMatcher is a value that can be passed to [Case.Puts] and has its own
// matching semantics.
type ValueMatcher interface{ matchValue(any) bool }

// Anything matches anything. It is useful when the value contains information
// that is useful when the test fails.
var Anything ValueMatcher = anything{}

type anything struct{}

func (anything) matchValue(any) bool { return true }

// ApproximatelyThreshold defines the threshold for matching float64 values when
// using [Approximately].
const ApproximatelyThreshold = 1e-15

// Approximately matches a float64 within the threshold defined by
// [ApproximatelyThreshold].
func Approximately(f float64) ValueMatcher { return approximately{f} }

type approximately struct{ value float64 }

func (a approximately) matchValue(value any) bool {
	if value, ok := value.(float64); ok {
		return matchFloat64(a.value, value, ApproximatelyThreshold)
	}
	return false
}

func matchFloat64(a, b, threshold float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 0) && math.IsInf(b, 0) &&
		math.Signbit(a) == math.Signbit(b) {
		return true
	}
	return math.Abs(a-b) <= threshold
}

// [StringMatching] matches any string matching a regexp pattern. If the pattern
// is not a valid regexp, the function panics.
func StringMatching(p string) ValueMatcher { return stringMatching{regexp.MustCompile(p)} }

type stringMatching struct{ pattern *regexp.Regexp }

func (s stringMatching) matchValue(value any) bool {
	if value, ok := value.(string); ok {
		return s.pattern.MatchString(value)
	}
	return false
}

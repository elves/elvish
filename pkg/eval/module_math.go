package eval

import "math"

// MathNs contains essential math functions.
var MathNs = NewNs().AddGoFns("math", map[string]interface{}{
	"abs":   math.Abs,
	"ceil":  math.Ceil,
	"floor": math.Floor,
	"round": math.Round,
})

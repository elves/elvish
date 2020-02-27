// Package math exposes functionality from Go's math package as an elvish
// module.
package maths

import (
	"math"

	"github.com/elves/elvish/pkg/eval"
)

//elvdoc:fn abs
//
// ```elvish
// math:abs $float64
// ```
//
// Compute the absolute value of a number. Example:
//
// ```elvish-transcript
// ~> math:abs 1.2
// > (float64 1.2)
// ~> math:abs -5.3
// > (float64 5.3)
// ```

//elvdoc:fn ceil
//
// ```elvish
// math:ceil $float64
// ```
//
// Compute the floor value of a number. Example:
//
// ```elvish-transcript
// ~> math:ceil 1.1
// > (float64 2)
// ~> math:ceil -2.3
// > (float64 -2)
// ```

//elvdoc:fn cos
//
// ```elvish
// math:cos $float64
// ```
//
// Compute the cosine of a number in radians (not degress). Example:
//
// ```elvish-transcript
// ~> math:cos 0
// > (float64 1)
// ~> math:cos 3.14159265
// > (float64 -1)
// ```

//elvdoc:fn floor
//
// ```elvish
// math:floor $float64
// ```
//
// Compute the floor value of a number. Example:
//
// ```elvish-transcript
// ~> math:floor 1.1
// > (float64 1)
// ~> math:floor -2.3
// > (float64 -3)
// ```

//elvdoc:fn log
//
// ```elvish
// math:log $float64
// ```
//
// Compute the natural logarithm of a number. Example:
//
// ```elvish-transcript
// ~> math:log 1.0
// > (float64 1)
// ~> math:log -2.3
// > (float64 NaN)
// ```

//elvdoc:fn log10
//
// ```elvish
// math:log10 $float64
// ```
//
// Compute the base 10 logarithm of a number. Example:
//
// ```elvish-transcript
// ~> math:log10 100.0
// > (float64 2)
// ~> math:log10 -1.7
// > (float64 NaN)
// ```

//elvdoc:fn log2
//
// ```elvish
// math:log2 $float64
// ```
//
// Compute the base 2 logarithm of a number. Example:
//
// ```elvish-transcript
// ~> math:log2 8
// > (float64 3)
// ~> math:log2 -5.3
// > (float64 NaN)
// ```

//elvdoc:fn sin
//
// ```elvish
// math:sin $float64
// ```
//
// Compute the sine of a number in radians (not degress). Example:
//
// ```elvish-transcript
// ~> math:sin 0
// > (float64 0)
// ~> math:sin 3.14159265
// > (float64 3.5897930298416118e-09)
// ```

//elvdoc:fn tan
//
// ```elvish
// math:tan $float64
// ```
//
// Compute the tangent of a number in radians (not degress). Example:
//
// ```elvish-transcript
// ~> math:tan 0
// > (float64 0)
// ~> math:tan 3.14159265
// > (float64 -0.0000000035897930298416118)
// ```

// Ns is the namespace for the math: module.
var Ns = eval.NewNs().AddGoFns("math:", fns)

var fns = map[string]interface{}{
	"abs":   math.Abs,
	"ceil":  math.Ceil,
	"cos":   math.Cos,
	"floor": math.Floor,
	"log":   math.Log,
	"log10": math.Log10,
	"log2":  math.Log2,
	"round": math.Round,
	"sin":   math.Sin,
	"tan":   math.Tan,
}

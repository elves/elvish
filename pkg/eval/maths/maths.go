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

// Ns is the namespace for the math: module.
var Ns = eval.NewNs().AddGoFns("math:", fns)

var fns = map[string]interface{}{
	"abs":   math.Abs,
	"ceil":  math.Ceil,
	"floor": math.Floor,
	"round": math.Round,
}

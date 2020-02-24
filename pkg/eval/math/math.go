// Package math exposes functionality from Go's math package as an elvish
// module.
package math

import (
	"math"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vars"
)

//elvdoc:var e
//
// ```elvish
// $math:e
// ```
//
// The value of
// [`e`](https://en.wikipedia.org/wiki/E_(mathematical_constant)):
// 2.718281.... This variable is read-only.

//elvdoc:var pi
//
// ```elvish
// $math:pi
// ```
//
// The value of [`π`](https://en.wikipedia.org/wiki/Pi): 3.141592.... This
// variable is read-only.

//elvdoc:fn abs
//
// ```elvish
// math:abs $number
// ```
//
// Computes the absolute value `$number`. Examples:
//
// ```elvish-transcript
// ~> math:abs 1.2
// ▶ (float64 1.2)
// ~> math:abs -5.3
// ▶ (float64 5.3)
// ```

//elvdoc:fn ceil
//
// ```elvish
// math:ceil $number
// ```
//
// Computes the ceiling of `$number`.
// Read the [Go documentation](https://godoc.org/math#Ceil) for the details of
// how this behaves. Examples:
//
// ```elvish-transcript
// ~> math:ceil 1.1
// ▶ (float64 2)
// ~> math:ceil -2.3
// ▶ (float64 -2)
// ```

//elvdoc:fn cos
//
// ```elvish
// math:cos $number
// ```
//
// Computes the cosine of `$number` in units of radians (not degrees).
// Examples:
//
// ```elvish-transcript
// ~> math:cos 0
// ▶ (float64 1)
// ~> math:cos 3.14159265
// ▶ (float64 -1)
// ```

//elvdoc:fn floor
//
// ```elvish
// math:floor $number
// ```
//
// Computes the floor of `$number`.
// Read the [Go documentation](https://godoc.org/math#Floor) for the details of
// how this behaves. Examples:
//
// ```elvish-transcript
// ~> math:floor 1.1
// ▶ (float64 1)
// ~> math:floor -2.3
// ▶ (float64 -3)
// ```

//elvdoc:fn is-inf
//
// ```elvish
// math:is-inf &sign=0 $number
// ```
//
// Tests whether the number is infinity. If sign > 0, tests whether `$number`
// is positive infinity. If sign < 0, tests whether `$number` is negative
// infinity. If sign == 0, tests whether `$number` is either infinity.
//
// ```elvish-transcript
// ~> math:is-inf 123
// ▶ $false
// ~> math:is-inf inf
// ▶ $true
// ~> math:is-inf -inf
// ▶ $true
// ~> math:is-inf &sign=1 inf
// ▶ $true
// ~> math:is-inf &sign=-1 inf
// ▶ $false
// ~> math:is-inf &sign=-1 -inf
// ▶ $true
// ```

//elvdoc:fn is-nan
//
// ```elvish
// math:is-nan $number $sign
// ```
//
// Tests whether the number is a NaN (not-a-number).
//
// ```elvish-transcript
// ~> math:is-nan 123
// ▶ $false
// ~> math:is-nan (float64 inf)
// ▶ $false
// ~> math:is-nan (float64 nan)
// ▶ $true
// ```

//elvdoc:fn log
//
// ```elvish
// math:log $number
// ```
//
// Computes the natural (base *e*) logarithm of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:log 1.0
// ▶ (float64 1)
// ~> math:log -2.3
// ▶ (float64 NaN)
// ```

//elvdoc:fn log10
//
// ```elvish
// math:log10 $number
// ```
//
// Computes the base 10 logarithm of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:log10 100.0
// ▶ (float64 2)
// ~> math:log10 -1.7
// ▶ (float64 NaN)
// ```

//elvdoc:fn log2
//
// ```elvish
// math:log2 $number
// ```
//
// Computes the base 2 logarithm of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:log2 8
// ▶ (float64 3)
// ~> math:log2 -5.3
// ▶ (float64 NaN)
// ```

//elvdoc:fn round
//
// ```elvish
// math:round $number
// ```
//
// Outputs the nearest integer, rounding half away from zero.
//
// ```elvish-transcript
// ~> math:round -1.1
// ▶ (float64 -1)
// ~> math:round 2.5
// ▶ (float64 3)
// ```

//elvdoc:fn round-to-even
//
// ```elvish
// math:round-to-even $number
// ```
//
// Outputs the nearest integer, rounding ties to even. Examples:
//
// ```elvish-transcript
// ~> math:round-to-even -1.1
// ▶ (float64 -1)
// ~> math:round-to-even 2.5
// ▶ (float64 2)
// ```

//elvdoc:fn sin
//
// ```elvish
// math:sin $number
// ```
//
// Computes the sine of `$number` in units of radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:sin 0
// ▶ (float64 0)
// ~> math:sin 3.14159265
// ▶ (float64 3.5897930298416118e-09)
// ```

//elvdoc:fn sqrt
//
// ```elvish
// math:sqrt $number
// ```
//
// Computes the square-root of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:sqrt 0
// ▶ (float64 0)
// ~> math:sqrt 4
// ▶ (float64 2)
// ~> math:sqrt -4
// ▶ (float64 NaN)
// ```

//elvdoc:fn tan
//
// ```elvish
// math:tan $number
// ```
//
// Computes the tangent `$number` in units of radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:tan 0
// ▶ (float64 0)
// ~> math:tan 3.14159265
// ▶ (float64 -0.0000000035897930298416118)
// ```

//elvdoc:fn trunc
//
// ```elvish
// math:trunc $number
// ```
//
// Outputs the integer portion of `$number`.
//
// ```elvish-transcript
// ~> math:trunc -1.1
// ▶ (float64 -1)
// ~> math:trunc 2.5
// ▶ (float64 2)
// ```

// Ns is the namespace for the math: module.
var Ns = eval.Ns{
	"e":  vars.NewReadOnly(math.E),
	"pi": vars.NewReadOnly(math.Pi),
}.AddGoFns("math:", fns)

var fns = map[string]interface{}{
	"abs":           math.Abs,
	"ceil":          math.Ceil,
	"cos":           math.Cos,
	"floor":         math.Floor,
	"is-inf":        isInf,
	"is-nan":        math.IsNaN,
	"log":           math.Log,
	"log10":         math.Log10,
	"log2":          math.Log2,
	"round":         math.Round,
	"round-to-even": math.RoundToEven,
	"sin":           math.Sin,
	"sqrt":          math.Sqrt,
	"tan":           math.Tan,
	"trunc":         math.Trunc,
}

type isInfOpts struct{ Sign int }

func (opts *isInfOpts) SetDefaultOptions() { opts.Sign = 0 }

func isInf(opts isInfOpts, arg float64) bool {
	return math.IsInf(arg, opts.Sign)
}

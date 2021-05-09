// Package math exposes functionality from Go's math package as an elvish
// module.
package math

import (
	"math"
	"math/big"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

// Ns is the namespace for the math: module.
var Ns = eval.NsBuilder{
	"e":  vars.NewReadOnly(math.E),
	"pi": vars.NewReadOnly(math.Pi),
}.AddGoFns("math:", fns).Ns()

var fns = map[string]interface{}{
	"abs":           abs,
	"acos":          math.Acos,
	"acosh":         math.Acosh,
	"asin":          math.Asin,
	"asinh":         math.Asinh,
	"atan":          math.Atan,
	"atanh":         math.Atanh,
	"ceil":          math.Ceil, // TODO: Make exactness-preserving
	"cos":           math.Cos,
	"cosh":          math.Cosh,
	"floor":         math.Floor, // TODO: Make exactness-preserving
	"is-inf":        isInf,
	"is-nan":        isNaN,
	"log":           math.Log,
	"log10":         math.Log10,
	"log2":          math.Log2,
	"max":           max,
	"min":           min,
	"pow":           math.Pow,         // TODO: Make exactness-preserving for integer exponents
	"pow10":         math.Pow10,       // TODO: Make exactness-preserving for integer exponents
	"round":         math.Round,       // TODO: Make exactness-preserving
	"round-to-even": math.RoundToEven, // TODO: Make exactness-preserving
	"sin":           math.Sin,
	"sinh":          math.Sinh,
	"sqrt":          math.Sqrt,
	"tan":           math.Tan,
	"tanh":          math.Tanh,
	"trunc":         math.Trunc, // TODO: Make exactness-preserving
}

//elvdoc:var e
//
// ```elvish
// $math:e
// ```
//
// Approximate value of
// [`e`](https://en.wikipedia.org/wiki/E_(mathematical_constant)):
// 2.718281.... This variable is read-only.

//elvdoc:var pi
//
// ```elvish
// $math:pi
// ```
//
// Approximate value of [`π`](https://en.wikipedia.org/wiki/Pi): 3.141592.... This
// variable is read-only.

//elvdoc:fn abs
//
// ```elvish
// math:abs $number
// ```
//
// Computes the absolute value `$number`. This function is exactness-preserving.
// Examples:
//
// ```elvish-transcript
// ~> math:abs 2
// ▶ (num 2)
// ~> math:abs -2
// ▶ (num 2)
// ~> math:abs 10000000000000000000
// ▶ (num 10000000000000000000)
// ~> math:abs -10000000000000000000
// ▶ (num 10000000000000000000)
// ~> math:abs 1/2
// ▶ (num 1/2)
// ~> math:abs -1/2
// ▶ (num 1/2)
// ~> math:abs 1.23
// ▶ (num 1.23)
// ~> math:abs -1.23
// ▶ (num 1.23)
// ```

const (
	maxInt = int(^uint(0) >> 1)
	minInt = -maxInt - 1
)

var absMinInt = new(big.Int).Abs(big.NewInt(int64(minInt)))

func abs(n vals.Num) vals.Num {
	switch n := n.(type) {
	case int:
		if n < 0 {
			if n == minInt {
				return absMinInt
			}
			return -n
		}
		return n
	case *big.Int:
		if n.Sign() < 0 {
			return new(big.Int).Abs(n)
		}
		return n
	case *big.Rat:
		if n.Sign() < 0 {
			return new(big.Rat).Abs(n)
		}
		return n
	case float64:
		return math.Abs(n)
	default:
		panic("unreachable")
	}
}

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

//elvdoc:fn acos
//
// ```elvish
// math:acos $number
// ```
//
// Outputs the arccosine of `$number`, in radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:acos 1
// ▶ (float64 1)
// ~> math:acos 1.00001
// ▶ (float64 NaN)
// ```

//elvdoc:fn acosh
//
// ```elvish
// math:acosh $number
// ```
//
// Outputs the inverse hyperbolic cosine of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:acosh 1
// ▶ (float64 0)
// ~> math:acosh 0
// ▶ (float64 NaN)
// ```

//elvdoc:fn asin
//
// ```elvish
// math:asin $number
// ```
//
// Outputs the arcsine of `$number`, in radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:asin 0
// ▶ (float64 0)
// ~> math:asin 1
// ▶ (float64 1.5707963267948966)
// ~> math:asin 1.00001
// ▶ (float64 NaN)
// ```

//elvdoc:fn asinh
//
// ```elvish
// math:asinh $number
// ```
//
// Outputs the inverse hyperbolic sine of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:asinh 0
// ▶ (float64 0)
// ~> math:asinh inf
// ▶ (float64 +Inf)
// ```

//elvdoc:fn atan
//
// ```elvish
// math:atan $number
// ```
//
// Outputs the arctangent of `$number`, in radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:atan 0
// ▶ (float64 0)
// ~> math:atan $math:inf
// ▶ (float64 1.5707963267948966)
// ```

//elvdoc:fn atanh
//
// ```elvish
// math:atanh $number
// ```
//
// Outputs the inverse hyperbolic tangent of `$number`. Examples:
//
// ```elvish-transcript
// ~> math:atanh 0
// ▶ (float64 0)
// ~> math:atanh 1
// ▶ (float64 +Inf)
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

//elvdoc:fn cosh
//
// ```elvish
// math:cosh $number
// ```
//
// Computes the hyperbolic cosine of `$number`. Example:
//
// ```elvish-transcript
// ~> math:cosh 0
// ▶ (float64 1)
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

type isInfOpts struct{ Sign int }

func (opts *isInfOpts) SetDefaultOptions() { opts.Sign = 0 }

func isInf(opts isInfOpts, n vals.Num) bool {
	if f, ok := n.(float64); ok {
		return math.IsInf(f, opts.Sign)
	}
	return false
}

//elvdoc:fn is-nan
//
// ```elvish
// math:is-nan $number
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

func isNaN(n vals.Num) bool {
	if f, ok := n.(float64); ok {
		return math.IsNaN(f)
	}
	return false
}

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

//elvdoc:fn max
//
// ```elvish
// math:max $number...
// ```
//
// Outputs the maximum number in the arguments. If there are no arguments,
// an exception is thrown. If any number is NaN then NaN is output. This
// function is exactness-preserving.
//
// Examples:
//
// ```elvish-transcript
// ~> math:max 3 5 2
// ▶ (num 5)
// ~> math:max (range 100)
// ▶ (num 99)
// ~> math:max 1/2 1/3 2/3
// ▶ (num 2/3)
// ```

func max(rawNums ...vals.Num) (vals.Num, error) {
	if len(rawNums) == 0 {
		return nil, errs.ArityMismatch{
			What: "arguments here", ValidLow: 1, ValidHigh: -1, Actual: 0}
	}
	nums := vals.UnifyNums(rawNums, 0)
	switch nums := nums.(type) {
	case []int:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n < nums[i] {
				n = nums[i]
			}
		}
		return n, nil
	case []*big.Int:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n.Cmp(nums[i]) < 0 {
				n = nums[i]
			}
		}
		return n, nil
	case []*big.Rat:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n.Cmp(nums[i]) < 0 {
				n = nums[i]
			}
		}
		return n, nil
	case []float64:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			n = math.Max(n, nums[i])
		}
		return n, nil
	default:
		panic("unreachable")
	}
}

//elvdoc:fn min
//
// ```elvish
// math:min $number...
// ```
//
// Outputs the minimum number in the arguments. If there are no arguments
// an exception is thrown. If any number is NaN then NaN is output. This
// function is exactness-preserving.
//
// Examples:
//
// ```elvish-transcript
// ~> math:min
// Exception: arity mismatch: arguments here must be 1 or more values, but is 0 values
// [tty 17], line 1: math:min
// ~> math:min 3 5 2
// ▶ (num 2)
// ~> math:min 1/2 1/3 2/3
// ▶ (num 1/3)
// ```

func min(rawNums ...vals.Num) (vals.Num, error) {
	if len(rawNums) == 0 {
		return nil, errs.ArityMismatch{
			What: "arguments here", ValidLow: 1, ValidHigh: -1, Actual: 0}
	}
	nums := vals.UnifyNums(rawNums, 0)
	switch nums := nums.(type) {
	case []int:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n > nums[i] {
				n = nums[i]
			}
		}
		return n, nil
	case []*big.Int:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n.Cmp(nums[i]) > 0 {
				n = nums[i]
			}
		}
		return n, nil
	case []*big.Rat:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			if n.Cmp(nums[i]) > 0 {
				n = nums[i]
			}
		}
		return n, nil
	case []float64:
		n := nums[0]
		for i := 1; i < len(nums); i++ {
			n = math.Min(n, nums[i])
		}
		return n, nil
	default:
		panic("unreachable")
	}
}

//elvdoc:fn pow
//
// ```elvish
// math:pow $base $exponent
// ```
//
// Output the result of raising `$base` to the power of `$exponent`. Examples:
//
// ```elvish-transcript
// ~> math:pow 3 2
// ▶ (float64 9)
// ~> math:pow -2 2
// ▶ (float64 4)
// ```
//
// @cf math:pow10

//elvdoc:fn pow10
//
// ```elvish
// math:pow10 $exponent
// ```
//
// Output the result of raising ten to the power of `$exponent` which must be
// an integer. Note that `$exponent > 308` results in +Inf and `$exponent <
// -323` results in zero. Examples:
//
// ```elvish-transcript
// ~> math:pow10 2
// ▶ (float64 100)
// ~> math:pow10 -3
// ▶ (float64 0.001)
// ```
//
// @cf math:pow

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

//elvdoc:fn sinh
//
// ```elvish
// math:sinh $number
// ```
//
// Computes the hyperbolic sine of `$number`. Example:
//
// ```elvish-transcript
// ~> math:sinh 0
// ▶ (float64 0)
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
// Computes the tangent of `$number` in units of radians (not degrees). Examples:
//
// ```elvish-transcript
// ~> math:tan 0
// ▶ (float64 0)
// ~> math:tan 3.14159265
// ▶ (float64 -0.0000000035897930298416118)
// ```

//elvdoc:fn tanh
//
// ```elvish
// math:tanh $number
// ```
//
// Computes the hyperbolic tangent of `$number`. Example:
//
// ```elvish-transcript
// ~> math:tanh 0
// ▶ (float64 0)
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

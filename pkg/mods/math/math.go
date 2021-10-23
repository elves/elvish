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
var Ns = eval.BuildNsNamed("math").
	AddVars(map[string]vars.Var{
		"e":  vars.NewReadOnly(math.E),
		"pi": vars.NewReadOnly(math.Pi),
	}).
	AddGoFns(map[string]interface{}{
		"abs":           abs,
		"acos":          math.Acos,
		"acosh":         math.Acosh,
		"asin":          math.Asin,
		"asinh":         math.Asinh,
		"atan":          math.Atan,
		"atanh":         math.Atanh,
		"ceil":          ceil,
		"cos":           math.Cos,
		"cosh":          math.Cosh,
		"floor":         floor,
		"is-inf":        isInf,
		"is-nan":        isNaN,
		"log":           math.Log,
		"log10":         math.Log10,
		"log2":          math.Log2,
		"max":           max,
		"min":           min,
		"pow":           pow,
		"round":         round,
		"round-to-even": roundToEven,
		"sin":           math.Sin,
		"sinh":          math.Sinh,
		"sqrt":          math.Sqrt,
		"tan":           math.Tan,
		"tanh":          math.Tanh,
		"trunc":         trunc,
	}).Ns()

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

//elvdoc:fn ceil
//
// ```elvish
// math:ceil $number
// ```
//
// Computes the least integer greater than or equal to `$number`. This function
// is exactness-preserving.
//
// The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
// NaN are themselves.
//
// Examples:
//
// ```elvish-transcript
// ~> math:floor 1
// ▶ (num 1)
// ~> math:floor 3/2
// ▶ (num 1)
// ~> math:floor -3/2
// ▶ (num -2)
// ~> math:floor 1.1
// ▶ (num 1.0)
// ~> math:floor -1.1
// ▶ (num -2.0)
// ```

var (
	big1 = big.NewInt(1)
	big2 = big.NewInt(2)
)

func ceil(n vals.Num) vals.Num {
	return integerize(n,
		math.Ceil,
		func(n *big.Rat) *big.Int {
			q := new(big.Int).Div(n.Num(), n.Denom())
			return q.Add(q, big1)
		})
}

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
// Computes the greatest integer less than or equal to `$number`. This function
// is exactness-preserving.
//
// The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
// NaN are themselves.
//
// Examples:
//
// ```elvish-transcript
// ~> math:floor 1
// ▶ (num 1)
// ~> math:floor 3/2
// ▶ (num 1)
// ~> math:floor -3/2
// ▶ (num -2)
// ~> math:floor 1.1
// ▶ (num 1.0)
// ~> math:floor -1.1
// ▶ (num -2.0)
// ```

func floor(n vals.Num) vals.Num {
	return integerize(n,
		math.Floor,
		func(n *big.Rat) *big.Int {
			return new(big.Int).Div(n.Num(), n.Denom())
		})
}

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
		return nil, errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: 0}
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
// Exception: arity mismatch: arguments must be 1 or more values, but is 0 values
// [tty 17], line 1: math:min
// ~> math:min 3 5 2
// ▶ (num 2)
// ~> math:min 1/2 1/3 2/3
// ▶ (num 1/3)
// ```

func min(rawNums ...vals.Num) (vals.Num, error) {
	if len(rawNums) == 0 {
		return nil, errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: 0}
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
// Outputs the result of raising `$base` to the power of `$exponent`.
//
// This function produces an exact result when `$base` is exact and `$exponent`
// is an exact integer. Otherwise it produces an inexact result.
//
// Examples:
//
// ```elvish-transcript
// ~> math:pow 3 2
// ▶ (num 9)
// ~> math:pow -2 2
// ▶ (num 4)
// ~> math:pow 1/2 3
// ▶ (num 1/8)
// ~> math:pow 1/2 -3
// ▶ (num 8)
// ~> math:pow 9 1/2
// ▶ (num 3.0)
// ~> math:pow 12 1.1
// ▶ (num 15.38506624784179)
// ```

func pow(base, exp vals.Num) vals.Num {
	if isExact(base) && isExactInt(exp) {
		// Produce exact result
		switch exp {
		case 0:
			return 1
		case 1:
			return base
		case -1:
			return new(big.Rat).Inv(vals.PromoteToBigRat(base))
		}
		exp := vals.PromoteToBigInt(exp)
		if isExactInt(base) && exp.Sign() > 0 {
			base := vals.PromoteToBigInt(base)
			return new(big.Int).Exp(base, exp, nil)
		}
		base := vals.PromoteToBigRat(base)
		if exp.Sign() < 0 {
			base = new(big.Rat).Inv(base)
			exp = new(big.Int).Neg(exp)
		}
		return new(big.Rat).SetFrac(
			new(big.Int).Exp(base.Num(), exp, nil),
			new(big.Int).Exp(base.Denom(), exp, nil))
	}

	// Produce inexact result
	basef := vals.ConvertToFloat64(base)
	expf := vals.ConvertToFloat64(exp)
	return math.Pow(basef, expf)
}

func isExact(n vals.Num) bool {
	switch n.(type) {
	case int, *big.Int, *big.Rat:
		return true
	default:
		return false
	}
}

func isExactInt(n vals.Num) bool {
	switch n.(type) {
	case int, *big.Int:
		return true
	default:
		return false
	}
}

//elvdoc:fn round
//
// ```elvish
// math:round $number
// ```
//
// Outputs the nearest integer, rounding half away from zero. This function is
// exactness-preserving.
//
// The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
// NaN are themselves.
//
// Examples:
//
// ```elvish-transcript
// ~> math:round 2
// ▶ (num 2)
// ~> math:round 1/3
// ▶ (num 0)
// ~> math:round 1/2
// ▶ (num 1)
// ~> math:round 2/3
// ▶ (num 1)
// ~> math:round -1/3
// ▶ (num 0)
// ~> math:round -1/2
// ▶ (num -1)
// ~> math:round -2/3
// ▶ (num -1)
// ~> math:round 2.5
// ▶ (num 3.0)
// ```

func round(n vals.Num) vals.Num {
	return integerize(n,
		math.Round,
		func(n *big.Rat) *big.Int {
			q, m := new(big.Int).QuoRem(n.Num(), n.Denom(), new(big.Int))
			m = m.Mul(m, big2)
			if m.CmpAbs(n.Denom()) < 0 {
				return q
			}
			if n.Sign() < 0 {
				return q.Sub(q, big1)
			}
			return q.Add(q, big1)
		})
}

//elvdoc:fn round-to-even
//
// ```elvish
// math:round-to-even $number
// ```
//
// Outputs the nearest integer, rounding ties to even. This function is
// exactness-preserving.
//
// The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
// NaN are themselves.
//
// Examples:
//
// ```elvish-transcript
// ~> math:round-to-even 2
// ▶ (num 2)
// ~> math:round-to-even 1/2
// ▶ (num 0)
// ~> math:round-to-even 3/2
// ▶ (num 2)
// ~> math:round-to-even 5/2
// ▶ (num 2)
// ~> math:round-to-even -5/2
// ▶ (num -2)
// ~> math:round-to-even 2.5
// ▶ (num 2.0)
// ~> math:round-to-even 1.5
// ▶ (num 2.0)
// ```

func roundToEven(n vals.Num) vals.Num {
	return integerize(n,
		math.RoundToEven,
		func(n *big.Rat) *big.Int {
			q, m := new(big.Int).QuoRem(n.Num(), n.Denom(), new(big.Int))
			m = m.Mul(m, big2)
			if diff := m.CmpAbs(n.Denom()); diff < 0 || diff == 0 && q.Bit(0) == 0 {
				return q
			}
			if n.Sign() < 0 {
				return q.Sub(q, big1)
			}
			return q.Add(q, big1)
		})
}

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
// Outputs the integer portion of `$number`. This function is exactness-preserving.
//
// The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
// NaN are themselves.
//
// Examples:
//
// ```elvish-transcript
// ~> math:trunc 1
// ▶ (num 1)
// ~> math:trunc 3/2
// ▶ (num 1)
// ~> math:trunc 5/3
// ▶ (num 1)
// ~> math:trunc -3/2
// ▶ (num -1)
// ~> math:trunc -5/3
// ▶ (num -1)
// ~> math:trunc 1.7
// ▶ (num 1.0)
// ~> math:trunc -1.7
// ▶ (num -1.0)
// ```

func trunc(n vals.Num) vals.Num {
	return integerize(n,
		math.Trunc,
		func(n *big.Rat) *big.Int {
			return new(big.Int).Quo(n.Num(), n.Denom())
		})
}

func integerize(n vals.Num, fnFloat func(float64) float64, fnRat func(*big.Rat) *big.Int) vals.Num {
	switch n := n.(type) {
	case int:
		return n
	case *big.Int:
		return n
	case *big.Rat:
		if n.Denom().IsInt64() && n.Denom().Int64() == 1 {
			// Elvish always normalizes *big.Rat with a denominator of 1 to
			// *big.Int, but we still try to be defensive here.
			return n.Num()
		}
		return fnRat(n)
	case float64:
		return fnFloat(n)
	default:
		panic("unreachable")
	}
}

package eval

import (
	"math"
	"math/rand"

	"github.com/elves/elvish/pkg/eval/vals"
)

// Numerical operations.

// TODO(xiaq): Document float64.

//elvdoc:fn + - * / ^
//
// ```elvish
// + $summand...
// - $minuend $subtrahend...
// * $factor...
// / $dividend $divisor...
// ^ $number
// ```
//
// Basic arithmetic operations of adding, subtraction, multiplication, division
// and exponentiation respectively.
//
// All of them can take multiple arguments:
//
// ```elvish-transcript
// ~> + 2 5 7 # 2 + 5 + 7
// ▶ 14
// ~> - 2 5 7 # 2 - 5 - 7
// ▶ -10
// ~> * 2 5 7 # 2 * 5 * 7
// ▶ 70
// ~> / 2 5 7 # 2 / 5 / 7
// ▶ 0.05714285714285715
// ~> ^ 2 3
// ▶ 8
// ~> ^ 2 0.5
// ▶ 1.4142135623730951
// ~> ^ 2 2 3
// ▶ 256
// ```
//
// When given one element, they all output their sole argument (given that it is a
// valid number). When given no argument,
//
// -   `+` outputs 0, and `*` and `^` outputs 1. You can think that they have a
// "hidden" argument of 0 or 1, which does not alter their behaviors (in
// mathematical terms, 0 and 1 are
// [identity elements](https://en.wikipedia.org/wiki/Identity_element) of
// addition and multiplication, respectively).
//
// -   `-` throws an exception.
//
// -   `/` becomes a synonym for `cd /`, due to the implicit cd feature. (The
// implicit cd feature will probably change to avoid this oddity).
//
// -    `^` is right associative, so that `^ 2 2 3` is equivalent to
// `^ 2 (^ 2 3)`, not to `^ (^2 2) 3`.

//elvdoc:fn %
//
// ```elvish
// % $dividend $divisor
// ```
//
// Output the remainder after dividing `$dividend` by `$divisor`. Both must be
// integers. Example:
//
// ```elvish-transcript
// ~> % 23 7
// ▶ 2
// ```

//elvdoc:fn &lt; &lt;= == != &gt; &gt;=
//
// ```elvish
// <  $number... # less
// <= $number... # less or equal
// == $number... # equal
// != $number... # not equal
// >  $number... # greater
// >= $number... # greater or equal
// ```
//
// Number comparisons. All of them accept an arbitrary number of arguments:
//
// 1.  When given fewer than two arguments, all output `$true`.
//
// 2.  When given two arguments, output whether the two arguments satisfy the named
// relationship.
//
// 3.  When given more than two arguments, output whether every adjacent pair of
// numbers satisfy the named relationship.
//
// Examples:
//
// ```elvish-transcript
// ~> == 3 3.0
// ▶ $true
// ~> < 3 4
// ▶ $true
// ~> < 3 4 10
// ▶ $true
// ~> < 6 9 1
// ▶ $false
// ```
//
// As a consequence of rule 3, the `!=` command outputs `$true` as long as any
// _adjacent_ pair of numbers are not equal, even if some numbers that are not
// adjacent are equal:
//
// ```elvish-transcript
// ~> != 5 5 4
// ▶ $false
// ~> != 5 6 5
// ▶ $true
// ```

//elvdoc:fn rand
//
// ```elvish
// rand
// ```
//
// Output a pseudo-random number in the interval [0, 1). Example:
//
// ```elvish-transcript
// ~> rand
// ▶ 0.17843564133528436
// ```

//elvdoc:fn randint
//
// ```elvish
// randint $low $high
// ```
//
// Output a pseudo-random integer in the interval [$low, $high). Example:
//
// ```elvish-transcript
// ~> # Emulate dice
// randint 1 7
// ▶ 6
// ```

func init() {
	addBuiltinFns(map[string]interface{}{
		// Constructor
		"float64": toFloat64,

		// Comparison
		"<":  lt,
		"<=": le,
		"==": eqNum,
		"!=": ne,
		">":  gt,
		">=": ge,

		// Arithmetics
		"+": plus,
		"-": minus,
		"*": times,
		"/": slash,
		"^": power,
		"%": mod,

		// Random
		"rand":    rand.Float64,
		"randint": randint,
	})
}

func toFloat64(f float64) float64 {
	return f
}

func lt(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] < nums[i+1]) {
			return false
		}
	}
	return true
}

func le(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] <= nums[i+1]) {
			return false
		}
	}
	return true
}

func eqNum(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] == nums[i+1]) {
			return false
		}
	}
	return true
}

func ne(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] != nums[i+1]) {
			return false
		}
	}
	return true
}

func gt(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] > nums[i+1]) {
			return false
		}
	}
	return true
}

func ge(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] >= nums[i+1]) {
			return false
		}
	}
	return true
}

func plus(nums ...float64) float64 {
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	return sum
}

func minus(sum float64, nums ...float64) float64 {
	if len(nums) == 0 {
		// Unary -
		return -sum
	}
	for _, f := range nums {
		sum -= f
	}
	return sum
}

func times(nums ...float64) float64 {
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	return prod
}

func power(nums ...float64) float64 {
	if len(nums) == 0 {
		return 1.0
	}
	pow := nums[len(nums)-1]
	for i := len(nums)-2; i >= 0; i-- {
		pow = math.Pow(nums[i], pow)
	}
	return pow
}

func slash(fm *Frame, args ...float64) error {
	if len(args) == 0 {
		// cd /
		return fm.Chdir("/")
	}
	// Division
	divide(fm, args[0], args[1:]...)
	return nil
}

func divide(fm *Frame, prod float64, nums ...float64) {
	out := fm.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- vals.FromGo(prod)
}

func mod(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrArgs
	}
	return a % b, nil
}

func randint(low, high int) (int, error) {
	if low >= high {
		return 0, ErrArgs
	}
	return low + rand.Intn(high-low), nil
}

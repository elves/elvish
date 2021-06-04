package eval

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Numerical operations.

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

func init() {
	addBuiltinFns(map[string]interface{}{
		// Constructor
		"float64":   toFloat64,
		"num":       num,
		"exact-num": exactNum,

		// Comparison
		"<":  lt,
		"<=": le,
		"==": eqNum,
		"!=": ne,
		">":  gt,
		">=": ge,

		// Arithmetic
		"+": add,
		"-": sub,
		"*": mul,
		// Also handles cd /
		"/": slash,
		"%": rem,

		// Random
		"rand":    rand.Float64,
		"randint": randint,
	})
}

//elvdoc:fn num
//
// ```elvish
// num $string-or-number
// ```
//
// Constructs a [typed number](./language.html#number).
//
// If the argument is a string, this command outputs the typed number the
// argument represents, or raises an exception if the argument is not a valid
// representation of a number. If the argument is already a typed number, this
// command outputs it as is.
//
// This command is usually not needed for working with numbers; see the
// discussion of [numerical commands](#numerical-commands).
//
// Examples:
//
// ```elvish-transcript
// ~> num 10
// ▶ (num 10)
// ~> num 0x10
// ▶ (num 16)
// ~> num 1/12
// ▶ (num 1/12)
// ~> num 3.14
// ▶ (num 3.14)
// ~> num (num 10)
// ▶ (num 10)
// ```

func num(n vals.Num) vals.Num {
	// Conversion is actually handled in vals/conversion.go.
	return n
}

//elvdoc:fn exact-num
//
// ```elvish
// exact-num $string-or-number
// ```
//
// Coerces the argument to an exact number. If the argument is infinity or NaN,
// an exception is thrown.
//
// If the argument is a string, it is converted to a typed number first. If the
// argument is already an exact number, it is returned as is.
//
// Examples:
//
// ```elvish-transcript
// ~> exact-num (num 0.125)
// ▶ (num 1/8)
// ~> exact-num 0.125
// ▶ (num 1/8)
// ~> exact-num (num 1)
// ▶ (num 1)
// ```
//
// Beware that seemingly simple fractions that can't be represented precisely in
// binary can result in the denominator being a very large power of 2:
//
// ```elvish-transcript
// ~> exact-num 0.1
// ▶ (num 3602879701896397/36028797018963968)
// ```

func exactNum(n vals.Num) (vals.Num, error) {
	if f, ok := n.(float64); ok {
		r := new(big.Rat).SetFloat64(f)
		if r == nil {
			return nil, errs.BadValue{What: "argument here",
				Valid: "finite float", Actual: vals.ToString(f)}
		}
		return r, nil
	}
	return n, nil
}

//elvdoc:fn float64
//
// ```elvish
// float64 $string-or-number
// ```
//
// Constructs a floating-point number.
//
// This command is deprecated; use [`num`](#num) instead.

func toFloat64(f float64) float64 {
	return f
}

//elvdoc:fn &lt; &lt;= == != &gt; &gt;= {#num-cmp}
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

func lt(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a < b },
		func(a, b *big.Int) bool { return a.Cmp(b) < 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) < 0 },
		func(a, b float64) bool { return a < b })

}

func le(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a <= b },
		func(a, b *big.Int) bool { return a.Cmp(b) <= 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) <= 0 },
		func(a, b float64) bool { return a <= b })
}

func eqNum(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a == b },
		func(a, b *big.Int) bool { return a.Cmp(b) == 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) == 0 },
		func(a, b float64) bool { return a == b })
}

func ne(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a != b },
		func(a, b *big.Int) bool { return a.Cmp(b) != 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) != 0 },
		func(a, b float64) bool { return a != b })
}

func gt(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a > b },
		func(a, b *big.Int) bool { return a.Cmp(b) > 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) > 0 },
		func(a, b float64) bool { return a > b })
}

func ge(nums ...vals.Num) bool {
	return chainCompare(nums,
		func(a, b int) bool { return a >= b },
		func(a, b *big.Int) bool { return a.Cmp(b) >= 0 },
		func(a, b *big.Rat) bool { return a.Cmp(b) >= 0 },
		func(a, b float64) bool { return a >= b })
}

func chainCompare(nums []vals.Num,
	p1 func(a, b int) bool, p2 func(a, b *big.Int) bool,
	p3 func(a, b *big.Rat) bool, p4 func(a, b float64) bool) bool {

	for i := 0; i < len(nums)-1; i++ {
		var r bool
		a, b := vals.UnifyNums2(nums[i], nums[i+1], 0)
		switch a := a.(type) {
		case int:
			r = p1(a, b.(int))
		case *big.Int:
			r = p2(a, b.(*big.Int))
		case *big.Rat:
			r = p3(a, b.(*big.Rat))
		case float64:
			r = p4(a, b.(float64))
		}
		if !r {
			return false
		}
	}
	return true
}

//elvdoc:fn + {#add}
//
// ```elvish
// + $num...
// ```
//
// Outputs the sum of all arguments, or 0 when there are no arguments.
//
// This command is [exactness-preserving](#exactness-preserving).
//
// Examples:
//
// ```elvish-transcript
// ~> + 5 2 7
// ▶ (num 14)
// ~> + 1/2 1/3 1/4
// ▶ (num 13/12)
// ~> + 1/2 0.5
// ▶ (num 1.0)
// ```

func add(rawNums ...vals.Num) vals.Num {
	nums := vals.UnifyNums(rawNums, vals.BigInt)
	switch nums := nums.(type) {
	case []*big.Int:
		acc := big.NewInt(0)
		for _, num := range nums {
			acc.Add(acc, num)
		}
		return vals.NormalizeBigInt(acc)
	case []*big.Rat:
		acc := big.NewRat(0, 1)
		for _, num := range nums {
			acc.Add(acc, num)
		}
		return vals.NormalizeBigRat(acc)
	case []float64:
		acc := float64(0)
		for _, num := range nums {
			acc += num
		}
		return acc
	default:
		panic("unreachable")
	}
}

//elvdoc:fn - {#sub}
//
// ```elvish
// - $x-num $y-num...
// ```
//
// Outputs the result of substracting from `$x-num` all the `$y-num`s, working
// from left to right. When no `$y-num` is given, outputs the negation of
// `$x-num` instead (in other words, `- $x-num` is equivalent to `- 0 $x-num`).
//
// This command is [exactness-preserving](#exactness-preserving).
//
// Examples:
//
// ```elvish-transcript
// ~> - 5
// ▶ (num -5)
// ~> - 5 2
// ▶ (num 3)
// ~> - 5 2 7
// ▶ (num -4)
// ~> - 1/2 1/3
// ▶ (num 1/6)
// ~> - 1/2 0.3
// ▶ (num 0.2)
// ~> - 10
// ▶ (num -10)
// ```

func sub(rawNums ...vals.Num) (vals.Num, error) {
	if len(rawNums) == 0 {
		return nil, errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: 0}
	}

	nums := vals.UnifyNums(rawNums, vals.BigInt)
	switch nums := nums.(type) {
	case []*big.Int:
		acc := &big.Int{}
		if len(nums) == 1 {
			acc.Neg(nums[0])
			return acc, nil
		}
		acc.Set(nums[0])
		for _, num := range nums[1:] {
			acc.Sub(acc, num)
		}
		return acc, nil
	case []*big.Rat:
		acc := &big.Rat{}
		if len(nums) == 1 {
			acc.Neg(nums[0])
			return acc, nil
		}
		acc.Set(nums[0])
		for _, num := range nums[1:] {
			acc.Sub(acc, num)
		}
		return acc, nil
	case []float64:
		if len(nums) == 1 {
			return -nums[0], nil
		}
		acc := nums[0]
		for _, num := range nums[1:] {
			acc -= num
		}
		return acc, nil
	default:
		panic("unreachable")
	}
}

//elvdoc:fn * {#mul}
//
// ```elvish
// * $num...
// ```
//
// Outputs the product of all arguments, or 1 when there are no arguments.
//
// This command is [exactness-preserving](#exactness-preserving). Additionally,
// when any argument is exact 0 and no other argument is a floating-point
// infinity, the result is exact 0.
//
// Examples:
//
// ```elvish-transcript
// ~> * 2 5 7
// ▶ (num 70)
// ~> * 1/2 0.5
// ▶ (num 0.25)
// ~> * 0 0.5
// ▶ (num 0)
// ```

func mul(rawNums ...vals.Num) vals.Num {
	hasExact0 := false
	hasInf := false
	for _, num := range rawNums {
		if num == 0 {
			hasExact0 = true
		}
		if f, ok := num.(float64); ok && math.IsInf(f, 0) {
			hasInf = true
			break
		}
	}
	if hasExact0 && !hasInf {
		return 0
	}

	nums := vals.UnifyNums(rawNums, vals.BigInt)
	switch nums := nums.(type) {
	case []*big.Int:
		acc := big.NewInt(1)
		for _, num := range nums {
			acc.Mul(acc, num)
		}
		return vals.NormalizeBigInt(acc)
	case []*big.Rat:
		acc := big.NewRat(1, 1)
		for _, num := range nums {
			acc.Mul(acc, num)
		}
		return vals.NormalizeBigRat(acc)
	case []float64:
		acc := float64(1)
		for _, num := range nums {
			acc *= num
		}
		return acc
	default:
		panic("unreachable")
	}
}

//elvdoc:fn / {#div}
//
// ```elvish
// / $x-num $y-num...
// ```
//
// Outputs the result of dividing `$x-num` with all the `$y-num`s, working from
// left to right. When no `$y-num` is given, outputs the reciprocal of `$x-num`
// instead (in other words, `/ $y-num` is equivalent to `/ 1 $y-num`).
//
// Dividing by exact 0 raises an exception. Dividing by inexact 0 results with
// either infinity or NaN according to floating-point semantics.
//
// This command is [exactness-preserving](#exactness-preserving). Additionally,
// when `$x-num` is exact 0 and no `$y-num` is exact 0, the result is exact 0.
//
// Examples:
//
// ```elvish-transcript
// ~> / 2
// ▶ (num 1/2)
// ~> / 2.0
// ▶ (num 0.5)
// ~> / 10 5
// ▶ (num 2)
// ~> / 2 5
// ▶ (num 2/5)
// ~> / 2 5 7
// ▶ (num 2/35)
// ~> / 0 1.0
// ▶ (num 0)
// ~> / 2 0
// Exception: bad value: divisor must be number other than exact 0, but is exact 0
// [tty 6], line 1: / 2 0
// ~> / 2 0.0
// ▶ (num +Inf)
// ```
//
// When given no argument, this command is equivalent to `cd /`, due to the
// implicit cd feature. (The implicit cd feature will probably change to avoid
// this oddity).

func slash(fm *Frame, args ...vals.Num) error {
	if len(args) == 0 {
		// cd /
		return fm.Evaler.Chdir("/")
	}
	// Division
	result, err := div(args...)
	if err == nil {
		fm.OutputChan() <- vals.FromGo(result)
	}
	return err
}

// ErrDivideByZero is thrown when attempting to divide by zero.
var ErrDivideByZero = errs.BadValue{
	What: "divisor", Valid: "number other than exact 0", Actual: "exact 0"}

func div(rawNums ...vals.Num) (vals.Num, error) {
	for _, num := range rawNums[1:] {
		if num == 0 {
			return nil, ErrDivideByZero
		}
	}
	if rawNums[0] == 0 {
		return 0, nil
	}
	nums := vals.UnifyNums(rawNums, vals.BigRat)
	switch nums := nums.(type) {
	case []*big.Rat:
		acc := &big.Rat{}
		acc.Set(nums[0])
		if len(nums) == 1 {
			acc.Inv(acc)
			return acc, nil
		}
		for _, num := range nums[1:] {
			acc.Quo(acc, num)
		}
		return acc, nil
	case []float64:
		acc := nums[0]
		if len(nums) == 1 {
			return 1 / acc, nil
		}
		for _, num := range nums[1:] {
			acc /= num
		}
		return acc, nil
	default:
		panic("unreachable")
	}
}

//elvdoc:fn % {#rem}
//
// ```elvish
// % $x $y
// ```
//
// Output the remainder after dividing `$x` by `$y`. The result has the same
// sign as `$x`. Both must be integers that can represented in a machine word
// (this limit may be lifted in future).
//
// Examples:
//
// ```elvish-transcript
// ~> % 10 3
// ▶ 1
// ~> % -10 3
// ▶ -1
// ~> % 10 -3
// ▶ 1
// ```

func rem(a, b int) (int, error) {
	// TODO: Support other number types
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a % b, nil
}

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

var ErrRandIntArgs = errors.New("$low >= $high")

func randint(low, high int) (int, error) {
	if low >= high {
		return 0, errs.BadValue{What: "low value",
			Valid: fmt.Sprintf("less than %d", high), Actual: fmt.Sprintf("%d", low)}
	}
	return low + rand.Intn(high-low), nil
}

package eval

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Numerical operations.

func init() {
	addBuiltinFns(map[string]any{
		// Constructor
		"num":         num,
		"exact-num":   exactNum,
		"inexact-num": inexactNum,

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
		"rand":      rand.Float64,
		"randint":   randint,
		"-randseed": randseed,

		"range": rangeFn,
	})

}

func num(n vals.Num) vals.Num {
	// Conversion is actually handled in vals/conversion.go.
	return n
}

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

func inexactNum(f float64) float64 {
	return f
}

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

func slash(fm *Frame, args ...vals.Num) error {
	if len(args) == 0 {
		// cd /
		return fm.Evaler.Chdir("/")
	}
	// Division
	result, err := div(args...)
	if err != nil {
		return err
	}
	return fm.ValueOutput().Put(vals.FromGo(result))
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

func rem(a, b int) (int, error) {
	// TODO: Support other number types
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a % b, nil
}

func randint(args ...int) (int, error) {
	var low, high int
	switch len(args) {
	case 1:
		low, high = 0, args[0]
	case 2:
		low, high = args[0], args[1]
	default:
		return -1, errs.ArityMismatch{What: "arguments",
			ValidLow: 1, ValidHigh: 2, Actual: len(args)}
	}
	if high <= low {
		return 0, errs.BadValue{What: "high value",
			Valid: fmt.Sprint("larger than ", low), Actual: strconv.Itoa(high)}
	}
	return low + rand.Intn(high-low), nil
}

//lint:ignore SA1019 useful for getting deterministic behavior in Elvish code.
func randseed(x int) { rand.Seed(int64(x)) }

type rangeOpts struct{ Step vals.Num }

// TODO: The default value can only be used implicitly; passing "range
// &step=nil" results in an error.
func (o *rangeOpts) SetDefaultOptions() { o.Step = nil }

func rangeFn(fm *Frame, opts rangeOpts, args ...vals.Num) error {
	var rawNums []vals.Num
	switch len(args) {
	case 1:
		rawNums = []vals.Num{0, args[0]}
	case 2:
		rawNums = []vals.Num{args[0], args[1]}
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: 2, Actual: len(args)}
	}
	if opts.Step != nil {
		rawNums = append(rawNums, opts.Step)
	}
	nums := vals.UnifyNums(rawNums, vals.Int)

	out := fm.ValueOutput()

	switch nums := nums.(type) {
	case []int:
		return rangeBuiltinNum(nums, out)
	case []*big.Int:
		return rangeBigNum(nums, out, bigIntDesc)
	case []*big.Rat:
		return rangeBigNum(nums, out, bigRatDesc)
	case []float64:
		return rangeBuiltinNum(nums, out)
	default:
		panic("unreachable")
	}
}

type builtinNum interface{ int | float64 }

func rangeBuiltinNum[T builtinNum](nums []T, out ValueOutput) error {
	start, end := nums[0], nums[1]
	var step T
	if start <= end {
		if len(nums) == 3 {
			step = nums[2]
			if step <= 0 {
				return errs.BadValue{
					What: "step", Valid: "positive", Actual: vals.ToString(step)}
			}
		} else {
			step = 1
		}
		for cur := start; cur < end; cur += step {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			if cur+step <= cur {
				break
			}
		}
	} else {
		if len(nums) == 3 {
			step = nums[2]
			if step >= 0 {
				return errs.BadValue{
					What: "step", Valid: "negative", Actual: vals.ToString(step)}
			}
		} else {
			step = -1
		}
		for cur := start; cur > end; cur += step {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			if cur+step >= cur {
				break
			}
		}
	}
	return nil
}

type bigNum[T any] interface {
	Cmp(T) int
	Sign() int
	Add(T, T) T
}

type bigNumDesc[T any] struct {
	one     T
	negOne  T
	newZero func() T
}

var bigIntDesc = bigNumDesc[*big.Int]{
	one:     big.NewInt(1),
	negOne:  big.NewInt(-1),
	newZero: func() *big.Int { return &big.Int{} },
}

var bigRatDesc = bigNumDesc[*big.Rat]{
	one:     big.NewRat(1, 1),
	negOne:  big.NewRat(-1, 1),
	newZero: func() *big.Rat { return &big.Rat{} },
}

func rangeBigNum[T bigNum[T]](nums []T, out ValueOutput, d bigNumDesc[T]) error {
	start, end := nums[0], nums[1]
	var step T
	if start.Cmp(end) <= 0 {
		if len(nums) == 3 {
			step = nums[2]
			if step.Sign() <= 0 {
				return errs.BadValue{
					What: "step", Valid: "positive", Actual: vals.ToString(step)}
			}
		} else {
			step = d.one
		}
		var cur, next T
		for cur = start; cur.Cmp(end) < 0; cur = next {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			next = d.newZero()
			next.Add(cur, step)
			cur = next
		}
	} else {
		if len(nums) == 3 {
			step = nums[2]
			if step.Sign() >= 0 {
				return errs.BadValue{
					What: "step", Valid: "negative", Actual: vals.ToString(step)}
			}
		} else {
			step = d.negOne
		}
		var cur, next T
		for cur = start; cur.Cmp(end) > 0; cur = next {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			next = d.newZero()
			next.Add(cur, step)
			cur = next
		}
	}
	return nil
}

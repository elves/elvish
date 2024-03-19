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
	AddGoFns(map[string]any{
		"abs":           abs,
		"acos":          math.Acos,
		"acosh":         math.Acosh,
		"asin":          math.Asin,
		"asinh":         math.Asinh,
		"atan":          math.Atan,
		"atan2":         math.Atan2,
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

func floor(n vals.Num) vals.Num {
	return integerize(n,
		math.Floor,
		func(n *big.Rat) *big.Int {
			return new(big.Int).Div(n.Num(), n.Denom())
		})
}

type isInfOpts struct{ Sign int }

func (opts *isInfOpts) SetDefaultOptions() { opts.Sign = 0 }

func isInf(opts isInfOpts, n vals.Num) bool {
	if f, ok := n.(float64); ok {
		return math.IsInf(f, opts.Sign)
	}
	return false
}

func isNaN(n vals.Num) bool {
	if f, ok := n.(float64); ok {
		return math.IsNaN(f)
	}
	return false
}

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

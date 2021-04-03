package vals

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

type Num interface{}

type NumSlice interface{}

func ParseNum(s string) Num {
	b := []byte(s)
	if strings.ContainsRune(s, '/') {
		// Parse as big.Rat
		var r big.Rat
		if r.UnmarshalText(b) == nil {
			return &r
		}
		return nil
	}
	// Try parsing as big.Int
	var z big.Int
	if z.UnmarshalText(b) == nil {
		if z.IsInt64() {
			return z.Int64()
		}
		return &z
	}
	// Try parsing as float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return nil
}

type NumType uint8

// Precedence used for unifying number types.
const (
	Int64 NumType = iota
	BigInt
	BigRat
	Float64
)

func UnifyNums(nums []Num, typ NumType) NumSlice {
	for _, num := range nums {
		if t := getNumType(num); t > typ {
			typ = t
		}
	}
	switch typ {
	case Int64:
		unified := make([]int64, len(nums))
		for i, num := range nums {
			unified[i] = num.(int64)
		}
		return unified
	case BigInt:
		unified := make([]*big.Int, len(nums))
		for i, num := range nums {
			switch num := num.(type) {
			case int64:
				unified[i] = big.NewInt(num)
			case *big.Int:
				unified[i] = num
			default:
				panic("unreachable")
			}
		}
		return unified
	case BigRat:
		unified := make([]*big.Rat, len(nums))
		for i, num := range nums {
			switch num := num.(type) {
			case int64:
				unified[i] = big.NewRat(num, 1)
			case *big.Int:
				var r big.Rat
				r.SetInt(num)
				unified[i] = &r
			case *big.Rat:
				unified[i] = num
			default:
				panic("unreachable")
			}
		}
		return unified
	case Float64:
		unified := make([]float64, len(nums))
		for i, num := range nums {
			switch num := num.(type) {
			case int64:
				unified[i] = float64(num)
			case *big.Int:
				if num.IsInt64() {
					unified[i] = float64(num.Int64())
				} else {
					// Definitely won't fit in float64
					unified[i] = math.Inf(num.Sign())
				}
			case *big.Rat:
				unified[i], _ = num.Float64()
			case float64:
				unified[i] = num
			default:
				panic("unreachable")
			}
		}
		return unified
	default:
		panic("unreachable")
	}
}

func getNumType(n Num) NumType {
	switch n.(type) {
	case int64:
		return Int64
	case *big.Int:
		return BigInt
	case *big.Rat:
		return BigRat
	case float64:
		return Float64
	default:
		panic("invalid num type " + fmt.Sprintf("%T", n))
	}
}

func NormalizeNum(n Num) Num {
	switch n := n.(type) {
	case int64:
		return n
	case *big.Int:
		if n.IsInt64() {
			return n.Int64()
		}
		return n
	case *big.Rat:
		if n.IsInt() {
			n := n.Num()
			if n.IsInt64() {
				return n.Int64()
			}
			return n
		}
		return n
	case float64:
		return n
	default:
		panic("invalid num type" + fmt.Sprintf("%T", n))
	}
}

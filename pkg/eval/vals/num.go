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
	z := &big.Int{}
	if z.UnmarshalText(b) == nil {
		if i, ok := getFixInt(z); ok {
			return i
		}
		return z
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
	FixInt NumType = iota
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
	case FixInt:
		unified := make([]int, len(nums))
		for i, num := range nums {
			unified[i] = num.(int)
		}
		return unified
	case BigInt:
		unified := make([]*big.Int, len(nums))
		for i, num := range nums {
			switch num := num.(type) {
			case int:
				unified[i] = big.NewInt(int64(num))
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
			case int:
				unified[i] = big.NewRat(int64(num), 1)
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
			case int:
				unified[i] = float64(num)
			case *big.Int:
				if num.IsInt64() {
					// Might fit in float64
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
	case int:
		return FixInt
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
	case int:
		return n
	case *big.Int:
		if i, ok := getFixInt(n); ok {
			return i
		}
		return n
	case *big.Rat:
		if n.IsInt() {
			if i, ok := getFixInt(n.Num()); ok {
				return i
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

func getFixInt(z *big.Int) (int, bool) {
	// TODO: Use a more efficient implementation by examining z.Bits
	if z.IsInt64() {
		i64 := z.Int64()
		i := int(i64)
		if int64(i) == i64 {
			return i, true
		}
	}
	return -1, false
}

package vals

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Num is a stand-in type for int, *big.Int, *big.Rat or float64. This type
// doesn't offer type safety, but is useful as a marker.
type Num interface{}

// NumSlice is a stand-in type for []int, []*big.Int, []*big.Rat or []float64.
// This type doesn't offer type safety, but is useful as a marker.
type NumSlice interface{}

// ParseNum parses a string into a suitable number type. If the string does not
// represent a valid number, it returns nil.
func ParseNum(s string) Num {
	if strings.ContainsRune(s, '/') {
		// Parse as big.Rat
		if z, ok := new(big.Rat).SetString(s); ok {
			return NormalizeBigRat(z)
		}
		return nil
	}
	// Try parsing as big.Int
	if z, ok := new(big.Int).SetString(s, 0); ok {
		return NormalizeBigInt(z)
	}
	// Try parsing as float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return nil
}

// NumType represents a number type.
type NumType uint8

// Possible values for NumType, sorted in the order of implicit conversion
// (lower types can be implicitly converted to higher types).
const (
	FixInt NumType = iota
	BigInt
	BigRat
	Float64
)

// UnifyNums unifies the given slice of numbers into the same type, converting
// those with lower NumType to the higest NumType present in the slice. The typ
// argument can be used to force the minimum NumType.
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

// NormalizeBigInt converts a big.Int to an int if it is within the range of
// int. Otherwise it returns n as is.
func NormalizeBigInt(z *big.Int) Num {
	if i, ok := getFixInt(z); ok {
		return i
	}
	return z
}

// NormalizeBigRat converts a big.Rat to a big.Int (or an int if within the
// range) if its denominator is 1.
func NormalizeBigRat(z *big.Rat) Num {
	if z.IsInt() {
		n := z.Num()
		if i, ok := getFixInt(n); ok {
			return i
		}
		return n
	}
	return z
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

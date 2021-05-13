package vals

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Design notes:
//
// The choice and relationship of number types in Elvish is closely modelled
// after R6RS's numerical tower (with the omission of complex types for now). In
// fact, there is a 1:1 correspondence between number types in Elvish and a
// typical R6RS implementation (the list below uses Chez Scheme's terminology;
// see https://www.scheme.com/csug8/numeric.html):
//
// int      : fixnum
// *big.Int : bignum
// *big.Rat : ratnum
// float64  : flonum
//
// Similar to Chez Scheme, *big.Int is only used for representing integers
// outside the range of int, and *big.Rat is only used for representing
// non-integer rationals. Furthermore, *big.Rat values are always in simplest
// form (this is guaranteed by the math/big library). As a consequence, each
// number in Elvish only has a single unique representation.
//
// Note that the only machine-native integer type included in the system is int.
// This is done primarily for the uniqueness of representation for each number,
// but also for simplicity - the vast majority of Go functions that take
// machine-native integers take int. When there is a genuine need to work with
// other machine-native integer types, you may have to manually convert from and
// to *big.Int and check for the relevant range of integers.

// Num is a stand-in type for int, *big.Int, *big.Rat or float64. This type
// doesn't offer type safety, but is useful as a marker; for example, it is
// respected when parsing function arguments.
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
	Int NumType = iota
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
	case Int:
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
			unified[i] = convertToFloat64(num)
		}
		return unified
	default:
		panic("unreachable")
	}
}

func getNumType(n Num) NumType {
	switch n.(type) {
	case int:
		return Int
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

func convertToFloat64(num Num) float64 {
	switch num := num.(type) {
	case int:
		return float64(num)
	case *big.Int:
		if num.IsInt64() { // probably fits in a float64
			return float64(num.Int64())
		}
		return math.Inf(num.Sign()) // definitely won't fit in a float64
	case *big.Rat:
		f, _ := num.Float64()
		return f
	case float64:
		return num
	default:
		panic("unreachable")
	}
}

// NormalizeBigInt converts a big.Int to an int if it is within the range of
// int. Otherwise it returns n as is.
func NormalizeBigInt(z *big.Int) Num {
	if i, ok := getInt(z); ok {
		return i
	}
	return z
}

// NormalizeBigRat converts a big.Rat to a big.Int (or an int if within the
// range) if its denominator is 1.
func NormalizeBigRat(z *big.Rat) Num {
	if z.IsInt() {
		n := z.Num()
		if i, ok := getInt(n); ok {
			return i
		}
		return n
	}
	return z
}

func getInt(z *big.Int) (int, bool) {
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

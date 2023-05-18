package vals

import (
	"math"
	"math/big"
)

// Basic predicate commands.

// Ordering relationship between two Elvish values.
type Ordering uint8

// Possible Ordering values.
const (
	CmpLess Ordering = iota
	CmpEqual
	CmpMore
	CmpUncomparable
)

// Cmp compares two Elvish values and returns the ordering relationship between
// them.
func Cmp(a, b any) Ordering {
	switch a := a.(type) {
	case int, *big.Int, *big.Rat, float64:
		switch b.(type) {
		case int, *big.Int, *big.Rat, float64:
			a, b := UnifyNums2(a, b, 0)
			switch a := a.(type) {
			case int:
				return compareInt(a, b.(int))
			case *big.Int:
				return compareInt(a.Cmp(b.(*big.Int)), 0)
			case *big.Rat:
				return compareInt(a.Cmp(b.(*big.Rat)), 0)
			case float64:
				return compareFloat(a, b.(float64))
			default:
				panic("unreachable")
			}
		}
	case string:
		if b, ok := b.(string); ok {
			switch {
			case a == b:
				return CmpEqual
			case a < b:
				return CmpLess
			default: // a > b
				return CmpMore
			}
		}
	case List:
		if b, ok := b.(List); ok {
			aIt := a.Iterator()
			bIt := b.Iterator()
			for aIt.HasElem() && bIt.HasElem() {
				o := Cmp(aIt.Elem(), bIt.Elem())
				if o != CmpEqual {
					return o
				}
				aIt.Next()
				bIt.Next()
			}
			switch {
			case a.Len() == b.Len():
				return CmpEqual
			case a.Len() < b.Len():
				return CmpLess
			default: // a.Len() > b.Len()
				return CmpMore
			}
		}
	case bool:
		if b, ok := b.(bool); ok {
			switch {
			case a == b:
				return CmpEqual
			//lint:ignore S1002 using booleans as values, not conditions
			case a == false: // b == true is implicit
				return CmpLess
			default: // a == true && b == false
				return CmpMore
			}
		}
	}
	return CmpUncomparable
}

func compareInt(a, b int) Ordering {
	if a < b {
		return CmpLess
	} else if a > b {
		return CmpMore
	}
	return CmpEqual
}

func compareFloat(a, b float64) Ordering {
	// For the sake of ordering, NaN's are considered equal to each
	// other and smaller than all numbers
	switch {
	case math.IsNaN(a):
		if math.IsNaN(b) {
			return CmpEqual
		}
		return CmpLess
	case math.IsNaN(b):
		return CmpMore
	case a < b:
		return CmpLess
	case a > b:
		return CmpMore
	default: // a == b
		return CmpEqual
	}
}

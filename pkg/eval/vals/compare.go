package vals

import (
	"math"
	"math/big"
)

type ordering uint8

const (
	CmpLess ordering = iota
	CmpEqual
	CmpMore
	CmpUncomparable
)

// Cmp compares two arbitrary Elvish types and returns a value that indicates
// whether the values are not comparable or the first value is less, equal, or
// greater than the second value.
func Cmp(a, b any) ordering {
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
	// TODO: Figure out a sane solution for comparing maps.  Obviously the
	// definition of "less than" is ill-defined for maps but it would be nice if
	// we had a solution similar to that for the List case above. One solution
	// is to treat each map as an ordered list of [key value] pairs (i.e.,
	// sorted by their keys) and use the List comparison logic above. Yes, that
	// behavior is hard to document but has the virtue of providing an
	// unambiguous means of sorting maps.
	//
	// case Map:
	//   if _, ok := b.(Map); ok {
	//     // Fill in the blank.
	//   }
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

func compareInt(a, b int) ordering {
	if a < b {
		return CmpLess
	} else if a > b {
		return CmpMore
	}
	return CmpEqual
}

func compareFloat(a, b float64) ordering {
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

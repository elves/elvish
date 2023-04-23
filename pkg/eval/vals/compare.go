package vals

import (
	"fmt"
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
				return cmpInt(a, b.(int))
			case *big.Int:
				return cmpInt(a.Cmp(b.(*big.Int)), 0)
			case *big.Rat:
				return cmpInt(a.Cmp(b.(*big.Rat)), 0)
			case float64:
				return cmpFloat(a, b.(float64))
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

// CmpTypes compares two arbitrary Elvish types and returns a value that
// indicates whether the type of the first value is less, equal, or greater than
// the type of the second value. This allows creating a total ordering even when
// dealing with heterogeneous types.
//
// The ordering is bool < num < string < list < map.
func CmpTypes(a, b any) ordering {
	switch a.(type) {
	case bool:
		return CmpLess // the bool type is always less than any other type
	case int, *big.Int, *big.Rat, float64:
		switch b.(type) {
		case bool:
			return CmpMore
		default:
			return CmpLess
		}
	case string:
		switch b.(type) {
		case bool:
			return CmpMore
		case int, *big.Int, *big.Rat, float64:
			return CmpMore
		default:
			return CmpLess
		}
	case List:
		switch b.(type) {
		case bool:
			return CmpMore
		case int, *big.Int, *big.Rat, float64:
			return CmpMore
		case string:
			return CmpMore
		default:
			return CmpLess
		}
	case Map:
		return CmpMore
	case StructMap:
		return CmpMore
	case PseudoStructMap:
		return CmpMore
	}

	// This will only be executed if we have failed to properly handle all the
	// types exposed by Elvish.
	panic(fmt.Sprintf("uncomparable types %T and %T", a, b))
}

func cmpInt(a, b int) ordering {
	if a < b {
		return CmpLess
	} else if a > b {
		return CmpMore
	}
	return CmpEqual
}

func cmpFloat(a, b float64) ordering {
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

// CmpKVSlice is useful when we need to sort a set of key/value pairs on the
// key. This supports hetergeneous key types by explicitly sorting uncomparable
// key types to produce a total ordering.
type CmpKVSlice [][2]any

func (s CmpKVSlice) Len() int { return len(s) }

func (s CmpKVSlice) Less(i, j int) bool {
	o := Cmp(s[i][0], s[j][0]) // compare the "keys" (the first value of the slice)
	if o == CmpUncomparable {
		o = CmpTypes(s[i][0], s[j][0])
	}
	return o == CmpLess
}

func (s CmpKVSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

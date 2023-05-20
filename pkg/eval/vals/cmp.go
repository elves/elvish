package vals

import (
	"math"
	"math/big"
	"reflect"
)

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
// them. Cmp(a, b) returns CmpEqual iff Equal(a, b) is true or both a and b are
// NaNs.
func Cmp(a, b any) Ordering {
	return cmpInner(a, b, Cmp)
}

func cmpInner(a, b any, recurse func(a, b any) Ordering) Ordering {
	// Keep the branches in the same order as [Equal].
	switch a := a.(type) {
	case nil:
		if b == nil {
			return CmpEqual
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
	case int, *big.Int, *big.Rat, float64:
		switch b.(type) {
		case int, *big.Int, *big.Rat, float64:
			a, b := UnifyNums2(a, b, 0)
			switch a := a.(type) {
			case int:
				return compareBuiltin(a, b.(int))
			case *big.Int:
				return compareBuiltin(a.Cmp(b.(*big.Int)), 0)
			case *big.Rat:
				return compareBuiltin(a.Cmp(b.(*big.Rat)), 0)
			case float64:
				return compareFloat(a, b.(float64))
			default:
				panic("unreachable")
			}
		}
	case string:
		if b, ok := b.(string); ok {
			return compareBuiltin(a, b)
		}
	case List:
		if b, ok := b.(List); ok {
			aIt := a.Iterator()
			bIt := b.Iterator()
			for aIt.HasElem() && bIt.HasElem() {
				o := recurse(aIt.Elem(), bIt.Elem())
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
	default:
		if Equal(a, b) {
			return CmpEqual
		}
	}
	return CmpUncomparable
}

func compareBuiltin[T interface{ int | string }](a, b T) Ordering {
	switch {
	case a < b:
		return CmpLess
	case a > b:
		return CmpMore
	default:
		return CmpEqual
	}
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

// CmpTotal is similar to [Cmp], but uses an artificial total ordering to avoid
// returning [CmpUncomparable]:
//
//   - If a and b have different types, it compares their types instead. The
//     ordering of types is guaranteed to be consistent during one Elvish
//     session, but is otherwise undefined.
//
//   - If a and b have the same type but are considered uncomparable by [Cmp],
//     it returns [CmpEqual] instead of [CmpUncomparable].
//
// All the underlying Go types of Elvish's number type are considered the same
// type.
//
// This function is mainly useful for sorting Elvish values that are not
// considered comparable by [Cmp]. Using this function as a comparator groups
// values by their types and sorts types that are comparable.
func CmpTotal(a, b any) Ordering {
	if o := compareBuiltin(typeOf(a), typeOf(b)); o != CmpEqual {
		return o
	}
	if o := cmpInner(a, b, CmpTotal); o != CmpUncomparable {
		return o
	}
	return CmpEqual
}

// typeOf returns a string unique to each type which can be used to order the
// types in a predictable, easy to describe, manner. We could use vals.Kind()
// since that would also provide a predictable, consistent, ordering but this
// gives us more control over the ordering than using the raw type names.
//
// See also vals.Kind(). If this changes there is a good chance vals.Kind() will
// also need to change.
func typeOf(x any) string {
	switch x := x.(type) {
	case nil:
		return "a"
	case bool:
		return "b"
	case int, *big.Int, *big.Rat, float64:
		return "c"
	case string:
		return "d"
	case List:
		return "e"
	case Map, StructMap:
		return "f"
	case File:
		return "g"
	case Kinder:
		// Handle Elvish types which are typically pseudo maps that we want to
		// classify as a distinct type; e.g., exceptions.
		return "y" + x.Kind()
	default:
		// This should rarely be executed. It exists to handle Go types we might
		// expose that are not mapped to a fundamental Elvish type. We want
		// these to appear after the Elvish types above.
		return "z" + reflect.TypeOf(x).String()
	}
}

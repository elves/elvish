package eval

import (
	"math"
	"math/big"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Basic predicate commands.

func init() {
	addBuiltinFns(map[string]any{
		"bool":    vals.Bool,
		"not":     not,
		"is":      is,
		"eq":      eq,
		"not-eq":  notEq,
		"compare": compare,
	})
}

func not(v any) bool {
	return !vals.Bool(v)
}

func is(args ...any) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			return false
		}
	}
	return true
}

func eq(args ...any) bool {
	for i := 0; i+1 < len(args); i++ {
		if !vals.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

func notEq(args ...any) bool {
	for i := 0; i+1 < len(args); i++ {
		if vals.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

// ErrUncomparable is raised by the compare and order commands when inputs contain
// uncomparable values.
var ErrUncomparable = errs.BadValue{
	What:  `inputs to "compare" or "order"`,
	Valid: "comparable values", Actual: "uncomparable values"}

func compare(fm *Frame, a, b any) (int, error) {
	switch cmp(a, b) {
	case less:
		return -1, nil
	case equal:
		return 0, nil
	case more:
		return 1, nil
	default:
		return 0, ErrUncomparable
	}
}

type ordering uint8

const (
	less ordering = iota
	equal
	more
	uncomparable
)

func cmp(a, b any) ordering {
	switch a := a.(type) {
	case int, *big.Int, *big.Rat, float64:
		switch b.(type) {
		case int, *big.Int, *big.Rat, float64:
			a, b := vals.UnifyNums2(a, b, 0)
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
				return equal
			case a < b:
				return less
			default: // a > b
				return more
			}
		}
	case vals.List:
		if b, ok := b.(vals.List); ok {
			aIt := a.Iterator()
			bIt := b.Iterator()
			for aIt.HasElem() && bIt.HasElem() {
				o := cmp(aIt.Elem(), bIt.Elem())
				if o != equal {
					return o
				}
				aIt.Next()
				bIt.Next()
			}
			switch {
			case a.Len() == b.Len():
				return equal
			case a.Len() < b.Len():
				return less
			default: // a.Len() > b.Len()
				return more
			}
		}
	case bool:
		if b, ok := b.(bool); ok {
			switch {
			case a == b:
				return equal
			//lint:ignore S1002 using booleans as values, not conditions
			case a == false: // b == true is implicit
				return less
			default: // a == true && b == false
				return more
			}
		}
	}
	return uncomparable
}

func compareInt(a, b int) ordering {
	if a < b {
		return less
	} else if a > b {
		return more
	}
	return equal
}

func compareFloat(a, b float64) ordering {
	// For the sake of ordering, NaN's are considered equal to each
	// other and smaller than all numbers
	switch {
	case math.IsNaN(a):
		if math.IsNaN(b) {
			return equal
		}
		return less
	case math.IsNaN(b):
		return more
	case a < b:
		return less
	case a > b:
		return more
	default: // a == b
		return equal
	}
}

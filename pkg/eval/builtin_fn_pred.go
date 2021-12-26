package eval

import (
	"math"
	"math/big"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Basic predicate commands.

//elvdoc:fn bool
//
// ```elvish
// bool $value
// ```
//
// Convert a value to boolean. In Elvish, only `$false` and errors are booleanly
// false. Everything else, including 0, empty strings and empty lists, is booleanly
// true:
//
// ```elvish-transcript
// ~> bool $true
// ▶ $true
// ~> bool $false
// ▶ $false
// ~> bool $ok
// ▶ $true
// ~> bool ?(fail haha)
// ▶ $false
// ~> bool ''
// ▶ $true
// ~> bool []
// ▶ $true
// ~> bool abc
// ▶ $true
// ```
//
// @cf not

func init() {
	addBuiltinFns(map[string]interface{}{
		"bool":    vals.Bool,
		"not":     not,
		"is":      is,
		"eq":      eq,
		"not-eq":  notEq,
		"compare": compare,
	})
}

//elvdoc:fn not
//
// ```elvish
// not $value
// ```
//
// Boolean negation. Examples:
//
// ```elvish-transcript
// ~> not $true
// ▶ $false
// ~> not $false
// ▶ $true
// ~> not $ok
// ▶ $false
// ~> not ?(fail error)
// ▶ $true
// ```
//
// **Note**: The related logical commands `and` and `or` are implemented as
// [special commands](language.html#special-commands) instead, since they do not
// always evaluate all their arguments. The `not` command always evaluates its
// only argument, and is thus a normal command.
//
// @cf bool

func not(v interface{}) bool {
	return !vals.Bool(v)
}

//elvdoc:fn is
//
// ```elvish
// is $values...
// ```
//
// Determine whether all `$value`s have the same identity. Writes `$true` when
// given no or one argument.
//
// The definition of identity is subject to change. Do not rely on its behavior.
//
// ```elvish-transcript
// ~> is a a
// ▶ $true
// ~> is a b
// ▶ $false
// ~> is [] []
// ▶ $true
// ~> is [a] [a]
// ▶ $false
// ```
//
// @cf eq
//
// Etymology: [Python](https://docs.python.org/3/reference/expressions.html#is).

func is(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			return false
		}
	}
	return true
}

//elvdoc:fn eq
//
// ```elvish
// eq $values...
// ```
//
// Determines whether all `$value`s are equal. Writes `$true` when
// given no or one argument.
//
// Two values are equal when they have the same type and value.
//
// For complex data structures like lists and maps, comparison is done
// recursively. A pseudo-map is equal to another pseudo-map with the same
// internal type (which is not exposed to Elvish code now) and value.
//
// ```elvish-transcript
// ~> eq a a
// ▶ $true
// ~> eq [a] [a]
// ▶ $true
// ~> eq [&k=v] [&k=v]
// ▶ $true
// ~> eq a [b]
// ▶ $false
// ```
//
// @cf is not-eq
//
// Etymology: [Perl](https://perldoc.perl.org/perlop.html#Equality-Operators).

func eq(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if !vals.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

//elvdoc:fn not-eq
//
// ```elvish
// not-eq $values...
// ```
//
// Determines whether every adjacent pair of `$value`s are not equal. Note that
// this does not imply that `$value`s are all distinct. Examples:
//
// ```elvish-transcript
// ~> not-eq 1 2 3
// ▶ $true
// ~> not-eq 1 2 1
// ▶ $true
// ~> not-eq 1 1 2
// ▶ $false
// ```
//
// @cf eq

func notEq(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if vals.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

//elvdoc:fn compare
//
// ```elvish
// compare $a $b
// ```
//
// Outputs -1 if `$a` < `$b`, 0 if `$a` = `$b`, and 1 if `$a` > `$b`.
//
// The following comparison algorithm is used:
//
// - Typed numbers are compared numerically. The comparison is consistent with
//   the [number comparison commands](#num-cmp), except that `NaN` values are
//   considered equal to each other and smaller than all other numbers.
//
// - Strings are compared lexicographically by bytes, consistent with the
//   [string comparison commands](#str-cmp). For UTF-8 encoded strings, this is
//   equivalent to comparing by codepoints.
//
// - Lists are compared lexicographically by elements, if the elements at the
//   same positions are comparable.
//
// If the ordering between two elements is not defined by the conditions above,
// i.e. if the value of `$a` or `$b` is not covered by any of the cases above or
// if they belong to different cases, a "bad value" exception is thrown.
//
// Examples:
//
// ```elvish-transcript
// ~> compare a b
// ▶ (num 1)
// ~> compare b a
// ▶ (num -1)
// ~> compare x x
// ▶ (num 0)
// ~> compare (float64 10) (float64 1)
// ▶ (num 1)
// ```
//
// Beware that strings that look like numbers are treated as strings, not
// numbers.
//
// @cf order

// ErrUncomparable is raised by the compare and order commands when inputs contain
// uncomparable values.
var ErrUncomparable = errs.BadValue{
	What:  `inputs to "compare" or "order"`,
	Valid: "comparable values", Actual: "uncomparable values"}

func compare(fm *Frame, a, b interface{}) (int, error) {
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

func cmp(a, b interface{}) ordering {
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

package eval

import (
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
	switch vals.Cmp(a, b) {
	case vals.CmpLess:
		return -1, nil
	case vals.CmpEqual:
		return 0, nil
	case vals.CmpMore:
		return 1, nil
	default:
		return 0, ErrUncomparable
	}
}

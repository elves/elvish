package eval

import "github.com/elves/elvish/eval/types"

// Basic predicate commands.

func init() {
	addBuiltinFns(map[string]interface{}{
		"bool":   types.Bool,
		"not":    not,
		"is":     is,
		"eq":     eq,
		"not-eq": notEq,
	})
}

func not(v interface{}) bool {
	return !types.Bool(v)
}

func is(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			return false
		}
	}
	return true
}

func eq(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if !types.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

func notEq(args ...interface{}) bool {
	for i := 0; i+1 < len(args); i++ {
		if types.Equal(args[i], args[i+1]) {
			return false
		}
	}
	return true
}

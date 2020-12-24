package eval

import "github.com/elves/elvish/pkg/eval/vals"

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
		"bool":   vals.Bool,
		"not":    not,
		"is":     is,
		"eq":     eq,
		"not-eq": notEq,
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
// **Note**: `not` is a regular command, and thus can be overriden by a
// function definition (though you shouldn't do so), while `and` and `or` are
// implemented as [special commands](language.html#special-commands).
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
// Determine whether all `$value`s are structurally equivalent. Writes `$true` when
// given no or one argument.
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

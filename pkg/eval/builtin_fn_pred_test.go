package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
)

func TestBool(t *testing.T) {
	Test(t,
		That(`bool $true`).Puts(true),
		That(`bool a`).Puts(true),
		That(`bool [a]`).Puts(true),
		// "Empty" values are also true in Elvish
		That(`bool []`).Puts(true),
		That(`bool [&]`).Puts(true),
		That(`bool 0`).Puts(true),
		That(`bool ""`).Puts(true),
		// Only errors and $false are false
		That(`bool ?(fail x)`).Puts(false),
		That(`bool $false`).Puts(false),
	)
}

func TestNot(t *testing.T) {
	Test(t,
		That(`not $false`).Puts(true),
		That(`not ?(fail x)`).Puts(true),
		That(`not $true`).Puts(false),
		That(`not 0`).Puts(false),
	)
}

func TestIs(t *testing.T) {
	Test(t,
		That(`is 1 1`).Puts(true),
		That(`is a b`).Puts(false),
		That(`is [] []`).Puts(true),
		That(`is [1] [1]`).Puts(false),
	)
}

func TestEq(t *testing.T) {
	Test(t,
		That(`eq 1 1`).Puts(true),
		That(`eq a b`).Puts(false),
		That(`eq [] []`).Puts(true),
		That(`eq [1] [1]`).Puts(true),
		That(`eq 1 1 2`).Puts(false),
	)
}

func TestNotEq(t *testing.T) {
	Test(t,
		That(`not-eq a b`).Puts(true),
		That(`not-eq a a`).Puts(false),
		// not-eq is true as long as each adjacent pair is not equal.
		That(`not-eq 1 2 1`).Puts(true),
	)
}

func TestCompare(t *testing.T) {
	Test(t,
		// Comparing strings.
		That("compare a b").Puts(-1),
		That("compare b a").Puts(1),
		That("compare x x").Puts(0),

		// Comparing numbers.
		That("compare (num 1) (num 2)").Puts(-1),
		That("compare (num 2) (num 1)").Puts(1),
		That("compare (num 3) (num 3)").Puts(0),

		That("compare (num 1/4) (num 1/2)").Puts(-1),
		That("compare (num 1/3) (num 0.2)").Puts(1),
		That("compare (num 3.0) (num 3)").Puts(0),

		That("compare (num nan) (num 3)").Puts(-1),
		That("compare (num 3) (num nan)").Puts(1),
		That("compare (num nan) (num nan)").Puts(0),

		// Comparing booleans.
		That("compare $true $false").Puts(1),
		That("compare $false $true").Puts(-1),
		That("compare $false $false").Puts(0),
		That("compare $true $true").Puts(0),

		// Comparing lists.
		That("compare [a, b] [a, a]").Puts(1),
		That("compare [a, a] [a, b]").Puts(-1),
		That("compare [x, y] [x, y]").Puts(0),

		// Different types are uncomparable without &total.
		That("compare 1 (num 1)").Throws(ErrUncomparable),
		That("compare x [x]").Throws(ErrUncomparable),
		That("compare a [&a=x]").Throws(ErrUncomparable),

		// Uncomparable types.
		That("compare { nop 1 } { nop 2}").Throws(ErrUncomparable),
		That("compare [&foo=bar] [&a=b]").Throws(ErrUncomparable),

		// Total ordering - different underlying number types are considered the
		// same type.
		That("compare &total (num 1) (num 3/2)").Puts(-1),
		That("compare &total (num 3/2) (num 2)").Puts(-1),

		// Total ordering - different types.
		That("== (compare &total foo (num 2)) (compare &total bar (num 10))").Puts(true),
		That("+ (compare &total foo (num 2)) (compare &total (num 2) foo)").Puts(0),

		// Total ordering - same uncomparable type.
		That("compare &total { nop 1 } { nop 2 }").Puts(0),
		That("compare &total [&foo=bar] [&a=b]").Puts(0),
	)
}

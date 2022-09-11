package eval_test

import (
	"math"
	"math/big"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestAll(t *testing.T) {
	Test(t,
		That(`put foo bar | all`).Puts("foo", "bar"),
		That(`echo foobar | all`).Puts("foobar"),
		That(`all [foo bar]`).Puts("foo", "bar"),
		thatOutputErrorIsBubbled("all [foo bar]"),
	)
}

func TestOne(t *testing.T) {
	Test(t,
		That(`put foo | one`).Puts("foo"),
		That(`put | one`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`put foo bar | one`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`one [foo]`).Puts("foo"),
		That(`one []`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`one [foo bar]`).Throws(ErrorWithType(errs.ArityMismatch{})),
		thatOutputErrorIsBubbled("one [foo]"),
	)
}

func TestTake(t *testing.T) {
	Test(t,
		That(`range 100 | take 2`).Puts(0, 1),
		thatOutputErrorIsBubbled("take 1 [foo bar]"),
	)
}

func TestDrop(t *testing.T) {
	Test(t,
		That(`range 100 | drop 98`).Puts(98, 99),
		thatOutputErrorIsBubbled("drop 1 [foo bar lorem]"),
	)
}

func TestCompact(t *testing.T) {
	Test(t,
		That(`put a a b b c | compact`).Puts("a", "b", "c"),
		That(`put a b a | compact`).Puts("a", "b", "a"),
		thatOutputErrorIsBubbled("compact [a a]"),
	)
}

func TestCount(t *testing.T) {
	Test(t,
		That(`range 100 | count`).Puts(100),
		That(`count [(range 100)]`).Puts(100),
		That(`count 123`).Puts(3),
		That(`count 1 2 3`).Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: 3},
			"count 1 2 3"),
		That(`count $true`).Throws(ErrorWithMessage("cannot get length of a bool")),
	)
}

func TestOrder(t *testing.T) {
	Test(t,
		// Ordering strings
		That("put foo bar ipsum | order").Puts("bar", "foo", "ipsum"),
		That("put foo bar bar | order").Puts("bar", "bar", "foo"),
		That("put 10 1 5 2 | order").Puts("1", "10", "2", "5"),

		// Ordering booleans
		That("put $true $false $true | order").Puts(false, true, true),
		That("put $false $true $false | order").Puts(false, false, true),

		// Ordering typed numbers
		// Only small integers
		That("put 10 1 1 | each $num~ | order").Puts(1, 1, 10),
		That("put 10 1 5 2 -1 | each $num~ | order").Puts(-1, 1, 2, 5, 10),
		// Small and large integers
		That("put 1 "+z+" 2 "+z+" | each $num~ | order").Puts(1, 2, bigInt(z), bigInt(z)),
		// Integers and rationals
		That("put 1 2 3/2 3/2 | each $num~ | order").
			Puts(1, big.NewRat(3, 2), big.NewRat(3, 2), 2),
		// Integers and floats
		That("put 1 1.5 2 1.5 | each $num~ | order").
			Puts(1, 1.5, 1.5, 2),
		// Mixed integers and floats.
		That("put (num 1) (num 1.5) (num 2) (num 1.5) | order").
			Puts(1, 1.5, 1.5, 2),
		// For the sake of ordering, NaN's are considered smaller than other numbers
		That("put NaN -1 NaN | each $num~ | order").Puts(math.NaN(), math.NaN(), -1),

		// Ordering lists
		That("put [b] [a] | order").Puts(vals.MakeList("a"), vals.MakeList("b")),
		That("put [a] [b] [a] | order").
			Puts(vals.MakeList("a"), vals.MakeList("a"), vals.MakeList("b")),
		That("put [(float64 10)] [(float64 2)] | order").
			Puts(vals.MakeList(2.0), vals.MakeList(10.0)),
		That("put [a b] [b b] [a c] | order").
			Puts(
				vals.MakeList("a", "b"),
				vals.MakeList("a", "c"), vals.MakeList("b", "b")),
		That("put [a] [] [a (num 2)] [a (num 1)] | order").
			Puts(vals.EmptyList, vals.MakeList("a"),
				vals.MakeList("a", 1), vals.MakeList("a", 2)),

		// Attempting to order uncomparable values
		That("put (num 1) 1 | order").
			Throws(ErrUncomparable, "order"),
		That("put 1 (num 1) | order").
			Throws(ErrUncomparable, "order"),
		That("put 1 (num 1) b | order").
			Throws(ErrUncomparable, "order"),
		That("put [a] a | order").
			Throws(ErrUncomparable, "order"),
		That("put [a] [(float64 1)] | order").
			Throws(ErrUncomparable, "order"),

		// &reverse
		That("put foo bar ipsum | order &reverse").Puts("ipsum", "foo", "bar"),

		// &less-than
		That("put 1 10 2 5 | order &less-than={|a b| < $a $b }").
			Puts("1", "2", "5", "10"),

		// &less-than writing more than one value
		That("put 1 10 2 5 | order &less-than={|a b| put $true $false }").
			Throws(
				errs.BadValue{
					What:  "output of the &less-than callback",
					Valid: "a single boolean", Actual: "2 values"},
				"order &less-than={|a b| put $true $false }"),

		// &less-than writing non-boolean value
		That("put 1 10 2 5 | order &less-than={|a b| put x }").
			Throws(
				errs.BadValue{
					What:  "output of the &less-than callback",
					Valid: "boolean", Actual: "string"},
				"order &less-than={|a b| put x }"),

		// &less-than throwing an exception
		That("put 1 10 2 5 | order &less-than={|a b| fail bad }").
			Throws(
				FailError{"bad"},
				"fail bad ", "order &less-than={|a b| fail bad }"),

		// &less-than and &reverse
		That("put 1 10 2 5 | order &reverse &less-than={|a b| < $a $b }").
			Puts("10", "5", "2", "1"),

		// Sort should be stable - test by pretending that all values but one
		// are equal, an check that the order among them has not changed.
		That("put l x o x r x e x m | order &less-than={|a b| eq $a x }").
			Puts("x", "x", "x", "x", "l", "o", "r", "e", "m"),

		thatOutputErrorIsBubbled("order [foo]"),
	)
}

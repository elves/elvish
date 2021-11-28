package eval_test

import (
	"math"
	"math/big"
	"testing"
	"unsafe"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestNsCmd(t *testing.T) {
	Test(t,
		That("put (ns [&name=value])[name]").Puts("value"),
		That("n: = (ns [&name=value]); put $n:name").Puts("value"),
		That("ns [&[]=[]]").Throws(errs.BadValue{
			What:  `key of argument of "ns"`,
			Valid: "string", Actual: "list"}),
	)
}

func TestMakeMap(t *testing.T) {
	Test(t,
		That("make-map []").Puts(vals.EmptyMap),
		That("make-map [[k v]]").Puts(vals.MakeMap("k", "v")),
		That("make-map [[k v] [k v2]]").Puts(vals.MakeMap("k", "v2")),
		That("make-map [[k1 v1] [k2 v2]]").
			Puts(vals.MakeMap("k1", "v1", "k2", "v2")),
		That("make-map [kv]").Puts(vals.MakeMap("k", "v")),
		That("make-map [{ } [k v]]").
			Throws(
				errs.BadValue{
					What: "input to make-map", Valid: "iterable", Actual: "fn"},
				"make-map [{ } [k v]]"),
		That("make-map [[k v] [k]]").
			Throws(
				errs.BadValue{
					What: "input to make-map", Valid: "iterable with 2 elements",
					Actual: "list with 1 elements"},
				"make-map [[k v] [k]]"),
	)
}

var maxInt = 1<<((unsafe.Sizeof(0)*8)-1) - 1
var maxDenseIntInFloat = float64(1 << 53)

func TestRange(t *testing.T) {
	Test(t,
		That("range 3").Puts(0, 1, 2),
		That("range 1 3").Puts(1, 2),
		That("range 0 10 &step=3").Puts(0, 3, 6, 9),
		// int overflow
		That("range &step=2 "+args(vals.ToString(maxInt-3), vals.ToString(maxInt))).
			Puts(maxInt-3, maxInt-1),
		// non-positive int step
		That("range &step=0 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "0"}),
		thatOutputErrorIsBubbled("range 1"),

		That("range "+z+" "+z3).Puts(bigInt(z), bigInt(z1), bigInt(z2)),
		That("range "+z+" "+z3+" &step=2").Puts(bigInt(z), bigInt(z2)),
		// non-positive bigint step
		That("range &step=-"+z+" 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-" + z}),
		thatOutputErrorIsBubbled("range "+z+" "+z1),

		That("range 23/10").Puts(0, 1, 2),
		That("range 1/10 23/10").Puts(
			big.NewRat(1, 10), big.NewRat(11, 10), big.NewRat(21, 10)),
		That("range 1/10 9/10 &step=3/10").Puts(
			big.NewRat(1, 10), big.NewRat(4, 10), big.NewRat(7, 10)),
		// non-positive bigrat step
		That("range &step=-1/2 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-1/2"}),
		thatOutputErrorIsBubbled("range 1/2 3/2"),

		That("range 1.2").Puts(0.0, 1.0),
		That("range &step=0.5 1 3").Puts(1.0, 1.5, 2.0, 2.5),
		// float64 overflow
		That("range "+args(vals.ToString(maxDenseIntInFloat-2), "+inf")).
			Puts(maxDenseIntInFloat-2, maxDenseIntInFloat-1, maxDenseIntInFloat),
		// non-positive float64 step
		That("range &step=-0.5 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-0.5"}),
		thatOutputErrorIsBubbled("range 1.2"),

		That("range").Throws(ErrorWithType(errs.ArityMismatch{})),
		That("range 0 1 2").Throws(ErrorWithType(errs.ArityMismatch{})),
	)
}

func TestRepeat(t *testing.T) {
	Test(t,
		That(`repeat 4 foo`).Puts("foo", "foo", "foo", "foo"),
		thatOutputErrorIsBubbled("repeat 1 foo"),
	)
}

func TestAssoc(t *testing.T) {
	Test(t,
		That(`put (assoc [0] 0 zero)[0]`).Puts("zero"),
		That(`put (assoc [&] k v)[k]`).Puts("v"),
		That(`put (assoc [&k=v] k v2)[k]`).Puts("v2"),
	)
}

func TestDissoc(t *testing.T) {
	Test(t,
		That(`has-key (dissoc [&k=v] k) k`).Puts(false),
		That("dissoc foo 0").Throws(ErrorWithMessage("cannot dissoc")),
	)
}

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
		That(`put | one`).Throws(AnyError),
		That(`put foo bar | one`).Throws(AnyError),
		That(`one [foo]`).Puts("foo"),
		That(`one []`).Throws(AnyError),
		That(`one [foo bar]`).Throws(AnyError),
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

func TestHasKey(t *testing.T) {
	Test(t,
		That(`has-key [foo bar] 0`).Puts(true),
		That(`has-key [foo bar] 0..1`).Puts(true),
		That(`has-key [foo bar] 0..20`).Puts(false),
		That(`has-key [&lorem=ipsum &foo=bar] lorem`).Puts(true),
		That(`has-key [&lorem=ipsum &foo=bar] loremwsq`).Puts(false),
	)
}

func TestHasValue(t *testing.T) {
	Test(t,
		That(`has-value [&lorem=ipsum &foo=bar] lorem`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] bar`).Puts(true),
		That(`has-value [foo bar] bar`).Puts(true),
		That(`has-value [foo bar] badehose`).Puts(false),
		That(`has-value "foo" o`).Puts(true),
		That(`has-value "foo" d`).Puts(false),
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

func TestKeys(t *testing.T) {
	Test(t,
		That(`keys [&]`).DoesNothing(),
		That(`keys [&a=foo]`).Puts("a"),
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		That(`keys [&a=foo &b=bar] | order`).Puts("a", "b"),
		That("keys (num 1)").Throws(ErrorWithMessage("cannot iterate keys of number")),
		thatOutputErrorIsBubbled("keys [&a=foo]"),
	)
}

func TestOrder(t *testing.T) {
	Test(t,
		// Ordering strings
		That("put foo bar ipsum | order").Puts("bar", "foo", "ipsum"),
		That("put foo bar bar | order").Puts("bar", "bar", "foo"),
		That("put 10 1 5 2 | order").Puts("1", "10", "2", "5"),

		// Ordering numbers
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
		That("put [a] [] [a (float64 2)] [a (float64 1)] | order").
			Puts(vals.EmptyList, vals.MakeList("a"),
				vals.MakeList("a", 1.0), vals.MakeList("a", 2.0)),

		// Attempting to order uncomparable values
		That("put a (float64 1) b (float64 2) | order").
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

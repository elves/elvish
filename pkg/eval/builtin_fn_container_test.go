package eval_test

import (
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
		That("var n: = (ns [&name=value]); put $n:name").Puts("value"),
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

var (
	maxInt = 1<<((unsafe.Sizeof(0)*8)-1) - 1
	minInt = -maxInt - 1

	maxDenseIntInFloat = float64(1 << 53)
)

func TestRange(t *testing.T) {
	Test(t,
		// Basic argument sanity checks.
		That("range").Throws(ErrorWithType(errs.ArityMismatch{})),
		That("range 0 1 2").Throws(ErrorWithType(errs.ArityMismatch{})),

		// Int count up.
		That("range 3").Puts(0, 1, 2),
		That("range 1 3").Puts(1, 2),
		// Int count down.
		That("range -1 10 &step=3").Puts(-1, 2, 5, 8),
		That("range 3 -3").Puts(3, 2, 1, 0, -1, -2),
		// Near maxInt or minInt.
		That("range "+args(maxInt-2, maxInt)).Puts(maxInt-2, maxInt-1),
		That("range "+args(maxInt, maxInt-2)).Puts(maxInt, maxInt-1),
		That("range "+args(minInt, minInt+2)).Puts(minInt, minInt+1),
		That("range "+args(minInt+2, minInt)).Puts(minInt+2, minInt+1),
		// Invalid step given the "start" and "end" values of the range.
		That("range &step=-1 1").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-1"}),
		That("range &step=1 1 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "1"}),
		thatOutputErrorIsBubbled("range 2"),

		// Big int count up.
		That("range "+z+" "+z3).Puts(bigInt(z), bigInt(z1), bigInt(z2)),
		That("range "+z+" "+z3+" &step=2").Puts(bigInt(z), bigInt(z2)),
		// Big int count down.
		That("range "+z3+" "+z).Puts(bigInt(z3), bigInt(z2), bigInt(z1)),
		That("range "+z3+" "+z+" &step=-2").Puts(bigInt(z3), bigInt(z1)),
		// Invalid big int step.
		That("range &step=-"+z+" 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-" + z}),
		That("range &step="+z+" 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: z}),
		thatOutputErrorIsBubbled("range "+z+" "+z1),

		// Rational count up.
		That("range 23/10").Puts(0, 1, 2),
		That("range 1/10 23/10").Puts(
			big.NewRat(1, 10), big.NewRat(11, 10), big.NewRat(21, 10)),
		That("range 23/10 1/10").Puts(
			big.NewRat(23, 10), big.NewRat(13, 10), big.NewRat(3, 10)),
		That("range 1/10 9/10 &step=3/10").Puts(
			big.NewRat(1, 10), big.NewRat(4, 10), big.NewRat(7, 10)),
		// Rational count down.
		That("range 9/10 0/10 &step=-3/10").Puts(
			big.NewRat(9, 10), big.NewRat(6, 10), big.NewRat(3, 10)),
		// Invalid rational step.
		That("range &step=-1/2 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-1/2"}),
		That("range &step=1/2 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "1/2"}),
		thatOutputErrorIsBubbled("range 1/2 3/2"),

		// Float64 count up.
		That("range 1.2").Puts(0.0, 1.0),
		That("range &step=0.5 1 3").Puts(1.0, 1.5, 2.0, 2.5),
		// Float64 count down.
		That("range 1.2 -1.2").Puts(1.2, Approximately{F: 0.2}, Approximately{F: -0.8}),
		That("range &step=-0.5 3 1").Puts(3.0, 2.5, 2.0, 1.5),
		// Near maxDenseIntInFloat.
		That("range "+args(maxDenseIntInFloat-2, "+inf")).
			Puts(maxDenseIntInFloat-2, maxDenseIntInFloat-1, maxDenseIntInFloat),
		That("range "+args(maxDenseIntInFloat, maxDenseIntInFloat-2)).
			Puts(maxDenseIntInFloat, maxDenseIntInFloat-1),
		// Invalid float64 step.
		That("range &step=-0.5 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-0.5"}),
		That("range &step=0.5 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "0.5"}),
		thatOutputErrorIsBubbled("range 1.2"),
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

		// Comparing lists.
		That("compare [a, b] [a, a]").Puts(1),
		That("compare [a, a] [a, b]").Puts(-1),
		That("compare [x, y] [x, y]").Puts(0),

		// Uncomparable values.
		That("compare 1 (num 1)").Throws(ErrUncomparable),
		That("compare x [x]").Throws(ErrUncomparable),
		That("compare a [&a=x]").Throws(ErrUncomparable),
	)
}

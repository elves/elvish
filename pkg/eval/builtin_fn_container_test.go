package eval_test

import (
	"math"
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/errs"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/eval/vals"
)

func TestBuiltinFnContainer(t *testing.T) {
	Test(t,
		That("put (ns [&name=value])[name]").Puts("value"),
		That("n: = (ns [&name=value]); put $n:name").Puts("value"),
		That("ns [&[]=[]]").Throws(errs.BadValue{
			What:  `key of argument of "ns"`,
			Valid: "string", Actual: "list"}),

		That("make-map []").Puts(vals.EmptyMap),
		That("make-map [[k v]]").Puts(vals.MakeMap("k", "v")),
		That("make-map [[k v] [k v2]]").Puts(vals.MakeMap("k", "v2")),
		That("make-map [[k1 v1] [k2 v2]]").
			Puts(vals.MakeMap("k1", "v1", "k2", "v2")),
		That("make-map [kv]").Puts(vals.MakeMap("k", "v")),
		That("make-map [{ }]").
			Throws(
				errs.BadValue{
					What: "input to make-map", Valid: "iterable", Actual: "fn"},
				"make-map [{ }]"),
		That("make-map [[k]]").
			Throws(
				errs.BadValue{
					What: "input to make-map", Valid: "iterable with 2 elements",
					Actual: "list with 1 elements"},
				"make-map [[k]]"),

		That(`range 3`).Puts(0.0, 1.0, 2.0),
		That(`range 1 3`).Puts(1.0, 2.0),
		That(`range 0 10 &step=3`).Puts(0.0, 3.0, 6.0, 9.0),

		That(`repeat 4 foo`).Puts("foo", "foo", "foo", "foo"),

		That(`put (assoc [0] 0 zero)[0]`).Puts("zero"),
		That(`put (assoc [&] k v)[k]`).Puts("v"),
		That(`put (assoc [&k=v] k v2)[k]`).Puts("v2"),
		That(`has-key (dissoc [&k=v] k) k`).Puts(false),

		That(`put foo bar | all`).Puts("foo", "bar"),
		That(`echo foobar | all`).Puts("foobar"),
		That(`all [foo bar]`).Puts("foo", "bar"),
		That(`put foo | one`).Puts("foo"),
		That(`put | one`).Throws(AnyError),
		That(`put foo bar | one`).Throws(AnyError),
		That(`one [foo]`).Puts("foo"),
		That(`one []`).Throws(AnyError),
		That(`one [foo bar]`).Throws(AnyError),

		That(`range 100 | take 2`).Puts(0.0, 1.0),
		That(`range 100 | drop 98`).Puts(98.0, 99.0),

		That(`has-key [foo bar] 0`).Puts(true),
		That(`has-key [foo bar] 0:1`).Puts(true),
		That(`has-key [foo bar] 0:20`).Puts(false),
		That(`has-key [&lorem=ipsum &foo=bar] lorem`).Puts(true),
		That(`has-key [&lorem=ipsum &foo=bar] loremwsq`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] lorem`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] bar`).Puts(true),
		That(`has-value [foo bar] bar`).Puts(true),
		That(`has-value [foo bar] badehose`).Puts(false),
		That(`has-value "foo" o`).Puts(true),
		That(`has-value "foo" d`).Puts(false),

		That(`range 100 | count`).Puts("100"),
		That(`count [(range 100)]`).Puts("100"),
		That(`count 123`).Puts("3"),
		That(`count 1 2 3`).Throws(
			errs.ArityMismatch{
				What: "arguments here", ValidLow: 0, ValidHigh: 1, Actual: 3},
			"count 1 2 3"),
		That(`count $true`).Throws(ErrorWithMessage("cannot get length of a bool")),

		That(`keys [&]`).DoesNothing(),
		That(`keys [&a=foo]`).Puts("a"),
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		That(`keys [&a=foo &b=bar] | order`).Puts("a", "b"),

		// Ordering strings
		That("put foo bar ipsum | order").Puts("bar", "foo", "ipsum"),
		That("put foo bar bar | order").Puts("bar", "bar", "foo"),
		That("put 10 1 5 2 | order").Puts("1", "10", "2", "5"),
		// Ordering numbers
		That("put 10 1 5 2 | each $float64~ | order").Puts(1.0, 2.0, 5.0, 10.0),
		That("put 10 1 1 | each $float64~ | order").Puts(1.0, 1.0, 10.0),
		That("put 10 NaN 1 | each $float64~ | order").Puts(math.NaN(), 1.0, 10.0),
		That("put NaN NaN 1 | each $float64~ | order").
			Puts(math.NaN(), math.NaN(), 1.0),
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
		That("put 1 10 2 5 | order &less-than=[a b]{ < $a $b }").
			Puts("1", "2", "5", "10"),
		// &less-than writing more than one value
		That("put 1 10 2 5 | order &less-than=[a b]{ put $true $false }").
			Throws(
				errs.BadValue{
					What:  "output of the &less-than callback",
					Valid: "a single boolean", Actual: "2 values"},
				"order &less-than=[a b]{ put $true $false }"),
		// &less-than writing non-boolean value
		That("put 1 10 2 5 | order &less-than=[a b]{ put x }").
			Throws(
				errs.BadValue{
					What:  "output of the &less-than callback",
					Valid: "boolean", Actual: "string"},
				"order &less-than=[a b]{ put x }"),
		// &less-than throwing an exception
		That("put 1 10 2 5 | order &less-than=[a b]{ fail bad }").
			Throws(
				FailError{"bad"},
				"fail bad ", "order &less-than=[a b]{ fail bad }"),
		// &less-than and &reverse
		That("put 1 10 2 5 | order &reverse &less-than=[a b]{ < $a $b }").
			Puts("10", "5", "2", "1"),
		// Sort should be stable - test by pretending that all values but one  d
		// are equal, an check that the order among them has not change        d
		That("put l x o x r x e x m | order &less-than=[a b]{ eq $a x }").
			Puts("x", "x", "x", "x", "l", "o", "r", "e", "m"),
	)
}

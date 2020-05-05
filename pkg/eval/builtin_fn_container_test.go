package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vals"
)

func TestBuiltinFnContainer(t *testing.T) {
	Test(t,
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
		That(`explode [foo bar]`).Puts("foo", "bar"),

		That(`put (assoc [0] 0 zero)[0]`).Puts("zero"),
		That(`put (assoc [&] k v)[k]`).Puts("v"),
		That(`put (assoc [&k=v] k v2)[k]`).Puts("v2"),
		That(`has-key (dissoc [&k=v] k) k`).Puts(false),

		That(`put foo bar | all`).Puts("foo", "bar"),
		That(`echo foobar | all`).Puts("foobar"),
		That(`all [foo bar]`).Puts("foo", "bar"),
		That(`put foo | one`).Puts("foo"),
		That(`put | one`).ThrowsAny(),
		That(`put foo bar | one`).ThrowsAny(),
		That(`one [foo]`).Puts("foo"),
		That(`one []`).ThrowsAny(),
		That(`one [foo bar]`).ThrowsAny(),

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
		That(`count $true`).ThrowsMessage("cannot get length of a bool"),

		That(`keys [&]`).DoesNothing(),
		That(`keys [&a=foo]`).Puts("a"),
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		// That(`keys [&a=foo &b=bar] | each echo | sort | each $put~`).Puts("a", "b"),
	)
}

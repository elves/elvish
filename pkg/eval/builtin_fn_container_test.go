package eval_test

import (
	"testing"

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

func TestConj(t *testing.T) {
	Test(t,
		That("conj [] a").Puts(vals.MakeList("a")),
		That("conj [a b]").Puts(vals.MakeList("a", "b")),
		That("conj [a b] c").Puts(vals.MakeList("a", "b", "c")),
		That("conj [a b] c d").Puts(vals.MakeList("a", "b", "c", "d")),
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
		That(`has-key [&k=v] k`).Puts(true),
		That(`has-key [&k=v] bad`).Puts(false),
		That(`has-key [lorem ipsum] 0`).Puts(true),
		That(`has-key [lorem ipsum] 0..`).Puts(true),
		That(`has-key [lorem ipsum] 0..=`).Puts(true),
		That(`has-key [lorem ipsum] ..2`).Puts(true),
		That(`has-key [lorem ipsum] ..=2`).Puts(false),
		That(`has-key [lorem ipsum] 2`).Puts(false),
		That(`has-key [lorem ipsum dolor sit] 0..4`).Puts(true),
		That(`has-key [lorem ipsum dolor sit] 0..=4`).Puts(false),
		That(`has-key [lorem ipsum dolor sit] 1..3`).Puts(true),
		That(`has-key [lorem ipsum dolor sit] 1..5`).Puts(false),
		That(`has-key [lorem ipsum dolor sit] -2..=-1`).Puts(true),
	)
}

func TestHasValue(t *testing.T) {
	Test(t,
		That(`has-value [&lorem=ipsum &foo=bar] lorem`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] bar`).Puts(true),
		That(`has-value [foo bar] bar`).Puts(true),
		That(`has-value [foo bar] badehose`).Puts(false),
		That(`has-value [[foo] [bar]] [foo]`).Puts(true),
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

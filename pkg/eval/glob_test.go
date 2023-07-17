package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestGlob_Simple(t *testing.T) {
	testutil.InTempDir(t)
	must.MkdirAll("z", "z2")
	must.CreateEmpty("bar", "foo", "ipsum", "lorem")

	Test(t,
		That("put *").Puts("bar", "foo", "ipsum", "lorem", "z", "z2"),
		That("put z*").Puts("z", "z2"),
		That("put ?").Puts("z"),
		That("put ????m").Puts("ipsum", "lorem"),
	)
}

func TestGlob_BraceConcat(t *testing.T) {
	testutil.InTempDir(t)
	must.CreateEmpty("xy.u", "xy.v", "xy.w", "xz.w")

	Test(t,
		That("put x*.{u v w}").Puts("xy.u", "xy.v", "xy.w", "xz.w"),
	)
}

func TestGlob_Recursive(t *testing.T) {
	testutil.InTempDir(t)
	must.MkdirAll("1/2/3")
	must.CreateEmpty("a.go", "1/a.go", "1/2/3/a.go")

	Test(t,
		That("put **").Puts("1/2/3/a.go", "1/2/3", "1/2", "1/a.go", "1", "a.go"),
		That("put **.go").Puts("1/2/3/a.go", "1/a.go", "a.go"),
		That("put 1**.go").Puts("1/2/3/a.go", "1/a.go"),
	)
}

func TestGlob_NoMatch(t *testing.T) {
	testutil.InTempDir(t)

	Test(t,
		That("put a/b/nonexistent*").Throws(ErrWildcardNoMatch),
		That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),
	)
}

func TestGlob_MatchHidden(t *testing.T) {
	testutil.InTempDir(t)
	must.MkdirAll("d", ".d")
	must.CreateEmpty("a", ".a", "d/a", "d/.a", ".d/a", ".d/.a")

	Test(t,
		That("put *").Puts("a", "d"),
		That("put *[match-hidden]").Puts(".a", ".d", "a", "d"),
		That("put *[match-hidden]/*").Puts(".d/a", "d/a"),
		That("put */*[match-hidden]").Puts("d/.a", "d/a"),
		That("put *[match-hidden]/*[match-hidden]").Puts(
			".d/.a", ".d/a", "d/.a", "d/a"),
	)
}

func TestGlob_RuneMatchers(t *testing.T) {
	testutil.InTempDir(t)
	must.CreateEmpty("a1", "a2", "b1", "c1", "ipsum", "lorem")

	Test(t,
		That("put *[letter]").Puts("ipsum", "lorem"),
		That("put ?[set:ab]*").Puts("a1", "a2", "b1"),
		That("put ?[range:a-c]*").Puts("a1", "a2", "b1", "c1"),
		That("put ?[range:a~c]*").Puts("a1", "a2", "b1"),
		That("put *[range:a-z]").Puts("ipsum", "lorem"),
		That("put *[range:a-zz]").Throws(ErrorWithMessage("bad range modifier: a-zz")),
		That("put *[range:foo]").Throws(ErrorWithMessage("bad range modifier: foo")),
	)
}

func TestGlob_But(t *testing.T) {
	testutil.InTempDir(t)
	must.CreateEmpty("bar", "foo", "ipsum", "lorem")

	Test(t,
		// Nonexistent files can also be excluded
		That("put *[but:foobar][but:ipsum]").Puts("bar", "foo", "lorem"),
	)
}

func TestGlob_Type(t *testing.T) {
	testutil.InTempDir(t)
	must.MkdirAll("d1", "d2", ".d", "b/c")
	must.CreateEmpty("bar", "foo", "ipsum", "lorem", "d1/f1", "d2/fm")

	Test(t,
		That("put **[type:dir]").Puts("b/c", "b", "d1", "d2"),
		That("put **[type:regular]m").Puts("d2/fm", "ipsum", "lorem"),
		That("put **[type:regular]f*").Puts("d1/f1", "d2/fm", "foo"),
		That("put **f*[type:regular]").Puts("d1/f1", "d2/fm", "foo"),

		That("put *[type:dir][type:regular]").Throws(ErrMultipleTypeModifiers),
		That("put **[type:dir]f*[type:regular]").Throws(ErrMultipleTypeModifiers),
		That("put **[type:unknown]").Throws(ErrUnknownTypeModifier),
	)
}

func TestGlob_BadOperation(t *testing.T) {
	testutil.InTempDir(t)

	Test(t,
		That("put *[[]]").Throws(ErrModifierMustBeString),
		That("put *[bad-mod]").Throws(ErrorWithMessage("unknown modifier bad-mod")),

		That("put *{ }").
			Throws(ErrorWithMessage("cannot concatenate glob-pattern and fn")),
		That("put { }*").
			Throws(ErrorWithMessage("cannot concatenate fn and glob-pattern")),
	)
}

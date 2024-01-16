package eval_test

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/must"

	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/testutil"
)

func TestKindOf(t *testing.T) {
	Test(t,
		That("kind-of a []").Puts("string", "list"),
		thatOutputErrorIsBubbled("kind-of a"),
	)
}

func TestConstantly(t *testing.T) {
	Test(t,
		That(`var f = (constantly foo); $f; $f`).Puts("foo", "foo"),
		thatOutputErrorIsBubbled("(constantly foo)"),
	)
}

func TestCallCommand(t *testing.T) {
	Test(t,
		That(`call {|arg &opt=v| put $arg $opt } [foo] [&opt=bar]`).
			Puts("foo", "bar"),
		That(`call { } [foo] [&[]=bar]`).
			Throws(errs.BadValue{What: "option key", Valid: "string", Actual: "list"}),
	)
}

func TestEval(t *testing.T) {
	Test(t,
		That("eval 'put x'").Puts("x"),
		// Using variable from the local scope.
		That("var x = foo; eval 'put $x'").Puts("foo"),
		// Setting a variable in the local scope.
		That("var x = foo; eval 'set x = bar'; put $x").Puts("bar"),
		// Using variable from the upvalue scope.
		That("var x = foo; { nop $x; eval 'put $x' }").Puts("foo"),
		// Specifying a namespace.
		That("var n = (ns [&x=foo]); eval 'put $x' &ns=$n").Puts("foo"),
		// Altering variables in the specified namespace.
		That("var n = (ns [&x=foo]); eval 'set x = bar' &ns=$n; put $n[x]").Puts("bar"),
		// Newly created variables do not appear in the local namespace.
		That("eval 'x = foo'; put $x").DoesNotCompile("variable $x not found"),
		// Newly created variables do not alter the specified namespace, either.
		That("var n = (ns [&]); eval &ns=$n 'var x = foo'; put $n[x]").
			Throws(vals.NoSuchKey("x"), "$n[x]"),
		// However, newly created variable can be accessed in the final
		// namespace using &on-end.
		That("eval &on-end={|n| put $n[x] } 'var x = foo'").Puts("foo"),
		// Parse error.
		That("eval '['").Throws(AnyParseError),
		// Compilation error.
		That("eval 'put $x'").Throws(ErrorWithType(&CompilationError{})),
		// Exception.
		That("eval 'fail x'").Throws(FailError{"x"}),
	)
}

func TestDeprecate(t *testing.T) {
	Test(t,
		That("deprecate msg").PrintsStderrWith("msg"),
		// Different call sites trigger multiple deprecation messages
		That("fn f { deprecate msg }", "f 2>"+os.DevNull, "f").
			PrintsStderrWith("msg"),
		// The same call site only triggers the message once
		That("fn f { deprecate msg}", "fn g { f }", "g 2>"+os.DevNull, "g 2>&1").
			DoesNothing(),
	)
}

func TestUseMod(t *testing.T) {
	testutil.InTempDir(t)
	must.WriteFile("mod.elv", "var x = value")

	Test(t,
		That("put (use-mod ./mod)[x]").Puts("value"),
	)
}

func TestResolve(t *testing.T) {
	libdir := testutil.InTempDir(t)
	must.WriteFile("mod.elv", "fn func { }")

	TestWithEvalerSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir} },
		That("resolve for").Puts("special"),
		That("resolve put").Puts("$put~"),
		That("fn f { }; resolve f").Puts("$f~"),
		That("use mod; resolve mod:func").Puts("$mod:func~"),
		That("resolve cat").Puts("(external cat)"),
		That(`resolve external`).Puts("$external~"),
	)
}

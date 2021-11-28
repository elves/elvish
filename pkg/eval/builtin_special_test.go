package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog/progtest"

	. "src.elv.sh/pkg/eval/evaltest"
	. "src.elv.sh/pkg/testutil"
)

func TestPragma(t *testing.T) {
	Test(t,
		That("pragma unknown-command").DoesNotCompile(),
		That("pragma unknown-command =").DoesNotCompile(),
		That("pragma unknown-command x").DoesNotCompile(),
		That("pragma bad-name = some-value").DoesNotCompile(),
		That("pragma unknown-command = bad").DoesNotCompile(),
	)
	// Actual effect of the unknown-command pragma is tested in TestCommand_External
}

func TestVar(t *testing.T) {
	Test(t,
		// Declaring one variable
		That("var x", "put $x").Puts(nil),
		// Declaring one variable whose name needs to be quoted
		That("var 'a/b'", "put $'a/b'").Puts(nil),
		// Declaring one variable whose name ends in ":".
		That("var a:").DoesNothing(),
		// Declaring a variable whose name ends in "~" initializes it to the
		// builtin nop function.
		That("var cmd~; cmd &ignored-opt ignored-arg").DoesNothing(),
		// Declaring multiple variables
		That("var x y", "put $x $y").Puts(nil, nil),
		// Declaring one variable with initial value
		That("var x = foo", "put $x").Puts("foo"),
		// Declaring multiple variables with initial values
		That("var x y = foo bar", "put $x $y").Puts("foo", "bar"),
		// Declaring multiple variables with initial values, including a rest
		// variable in the assignment LHS
		That("var x @y z = a b c d", "put $x $y $z").
			Puts("a", vals.MakeList("b", "c"), "d"),
		// An empty RHS is technically legal although rarely useful.
		That("var @x =", "put $x").Puts(vals.EmptyList),
		// Shadowing.
		That("var x = old; fn f { put $x }", "var x = new; put $x; f").
			Puts("new", "old"),
		// Explicit local: is allowed
		That("var local:x = foo", "put $x").Puts("foo"),

		// Variable name that must be quoted after $ must be quoted
		That("var a/b").DoesNotCompile(),
		// Multiple @ not allowed
		That("var x @y @z = a b c d").DoesNotCompile(),
		// Non-local not allowed
		That("var ns:a").DoesNotCompile(),
		// Index not allowed
		That("var a[0]").DoesNotCompile(),
		// Composite expression not allowed
		That("var a'b'").DoesNotCompile(),
	)
}

func TestSet(t *testing.T) {
	Test(t,
		// Setting one variable
		That("var x; set x = foo", "put $x").Puts("foo"),
		// An empty RHS is technically legal although rarely useful.
		That("var x; set @x =", "put $x").Puts(vals.EmptyList),
		// Variable must already exist
		That("set x = foo").DoesNotCompile(),
		// Not duplicating tests with TestCommand_Assignment.
		//
		// TODO: After legacy assignment form is removed, transfer tests here.

		// = is required.
		That("var x; set x").DoesNotCompile(),
	)
}

func TestDel(t *testing.T) {
	Setenv(t, "TEST_ENV", "test value")

	Test(t,
		// Deleting variable
		That("x = 1; del x").DoesNothing(),
		That("x = 1; del x; echo $x").DoesNotCompile(),
		That("x = 1; del :x; echo $x").DoesNotCompile(),
		That("x = 1; del local:x; echo $x").DoesNotCompile(),
		// Deleting environment variable
		That("has-env TEST_ENV", "del E:TEST_ENV", "has-env TEST_ENV").Puts(true, false),
		// Deleting variable whose name contains special characters
		That("'a/b' = foo; del 'a/b'").DoesNothing(),
		// Deleting element
		That("x = [&k=v &k2=v2]; del x[k2]; keys $x").Puts("k"),
		That("x = [[&k=v &k2=v2]]; del x[0][k2]; keys $x[0]").Puts("k"),

		// Error cases

		// Deleting nonexistent variable
		That("del x").DoesNotCompile(),
		// Deleting element of nonexistent variable
		That("del x[0]").DoesNotCompile(),
		// Deleting variable in non-local namespace
		That("var a: = (ns [&b=$nil])", "del a:b").DoesNotCompile(),
		// Variable name given with $
		That("x = 1; del $x").DoesNotCompile(),
		// Variable name not given as a single primary expression
		That("ab = 1; del a'b'").DoesNotCompile(),
		// Variable name not a string
		That("del [a]").DoesNotCompile(),
		// Variable name has sigil
		That("x = []; del @x").DoesNotCompile(),
		// Variable name not quoted when it should be
		That("'a/b' = foo; del a/b").DoesNotCompile(),

		// Index is multiple values
		That("x = [&k1=v1 &k2=v2]", "del x[k1 k2]").Throws(
			ErrorWithMessage("index must evaluate to a single value in argument to del"),
			"k1 k2"),
		// Index expression throws exception
		That("x = [&k]", "del x[(fail x)]").Throws(FailError{"x"}, "fail x"),
		// Value does not support element removal
		That("x = (num 1)", "del x[k]").Throws(
			ErrorWithMessage("value does not support element removal"),
			// TODO: Fix the stack trace so that it is "x[k]"
			"x[k"),
		// Intermediate element does not exist
		That("x = [&]", "del x[k][0]").Throws(
			ErrorWithMessage("no such key: k"),
			// TODO: Fix the stack trace so that it is "x[k]"
			"x"),
	)
}

func TestAnd(t *testing.T) {
	Test(t,
		That("and $true $false").Puts(false),
		That("and a b").Puts("b"),
		That("and $false b").Puts(false),
		That("and $true b").Puts("b"),
		// short circuit
		That("x = a; and $false (x = b); put $x").Puts(false, "a"),

		// Exception
		That("and a (fail x)").Throws(FailError{"x"}, "fail x"),
		thatOutputErrorIsBubbled("and a"),
	)
}

func TestOr(t *testing.T) {
	Test(t,
		That("or $true $false").Puts(true),
		That("or a b").Puts("a"),
		That("or $false b").Puts("b"),
		That("or $true b").Puts(true),
		// short circuit
		That("x = a; or $true (x = b); put $x").Puts(true, "a"),

		// Exception
		That("or $false (fail x)").Throws(FailError{"x"}, "fail x"),
		thatOutputErrorIsBubbled("or a"),
	)
}

func TestCoalesce(t *testing.T) {
	Test(t,
		That("coalesce a b").Puts("a"),
		That("coalesce $nil b").Puts("b"),
		That("coalesce $nil $nil").Puts(nil),
		That("coalesce").Puts(nil),
		// exception propagation
		That("coalesce $nil (fail foo)").Throws(FailError{"foo"}),
		// short circuit
		That("coalesce a (fail foo)").Puts("a"),

		thatOutputErrorIsBubbled("coalesce a"),
	)
}

func TestIf(t *testing.T) {
	Test(t,
		That("if true { put then }").Puts("then"),
		That("if $false { put then } else { put else }").Puts("else"),
		That("if $false { put 1 } elif $false { put 2 } else { put 3 }").
			Puts("3"),
		That("if $false { put 2 } elif true { put 2 } else { put 3 }").Puts("2"),

		// Exception in condition expression
		That("if (fail x) { }").Throws(FailError{"x"}, "fail x"),
	)
}

func TestTry(t *testing.T) {
	Test(t,
		That("try { nop } except { put bad } else { put good }").Puts("good"),
		That("try { e:false } except - { put bad } else { put good }").
			Puts("bad"),
		That("try { fail tr }").Throws(ErrorWithMessage("tr")),
		That("try { fail tr } finally { put final }").
			Puts("final").
			Throws(ErrorWithMessage("tr")),

		That("try { fail tr } except { fail ex } finally { put final }").
			Puts("final").
			Throws(ErrorWithMessage("ex")),

		That("try { fail tr } except { put ex } finally { fail final }").
			Puts("ex").
			Throws(ErrorWithMessage("final")),

		That("try { fail tr } except { fail ex } finally { fail final }").
			Throws(ErrorWithMessage("final")),

		// wrong syntax
		That("try { nop } except @a { }").DoesNotCompile(),

		// A quoted var name, that would be invalid as a bareword, should be allowed as the referent
		// in a `try...except...` block.
		That("try { fail hard } except 'x=' { put 'x= ='(to-string $'x=') }").
			Puts("x= =[&reason=[&content=hard &type=fail]]"),
	)
}

func TestWhile(t *testing.T) {
	Test(t,
		That("var x = (num 0)", "while (< $x 4) { put $x; set x = (+ $x 1) }").
			Puts(0, 1, 2, 3),
		// break
		That("var x = (num 0)", "while (< $x 4) { put $x; break }").Puts(0),
		// continue
		That("var x = (num 0)",
			"while (< $x 4) { put $x; set x = (+ $x 1); continue; put bad }").
			Puts(0, 1, 2, 3),
		// Exception in body
		That("var x = 0; while (< $x 4) { fail haha }").Throws(AnyError),
		// Exception in condition
		That("while (fail x) { }").Throws(FailError{"x"}, "fail x"),

		// else branch - not taken
		That("var x = 0; while (< $x 4) { put $x; set x = (+ $x 1) } else { put bad }").
			Puts("0", 1, 2, 3),
		// else branch - taken
		That("while $false { put bad } else { put good }").Puts("good"),
	)
}

func TestFor(t *testing.T) {
	Test(t,
		// for
		That("for x [tempora mores] { put 'O '$x }").
			Puts("O tempora", "O mores"),
		// break
		That("for x [a] { break } else { put $x }").DoesNothing(),
		// else
		That("for x [a] { put $x } else { put $x }").Puts("a"),
		// continue
		That("for x [a b] { put $x; continue; put $x; }").Puts("a", "b"),
		// else
		That("for x [] { } else { put else }").Puts("else"),
		That("for x [a] { } else { put else }").DoesNothing(),
		// Propagating exception.
		That("for x [a] { fail foo }").Throws(FailError{"foo"}),

		// More than one iterator.
		That("for {x,y} [] { }").DoesNotCompile(),
		// Invalid for loop lvalue. You can't use a var in a namespace other
		// than the local namespace as the lvalue in a for loop.
		That("for no-such-namespace:x [a b] { }").DoesNotCompile(),
		// Exception with the variable
		That("var a: = (ns [&])", "for a:b [] { }").Throws(
			ErrorWithMessage("no variable $a:b"),
			"a:b"),
		// Exception when evaluating iterable.
		That("for x [][0] { }").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		// More than one iterable.
		That("for x (put a b) { }").Throws(
			errs.ArityMismatch{What: "value being iterated",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"(put a b)"),
		// Non-iterable value
		That("for x (num 0) { }").Throws(ErrorWithMessage("cannot iterate number")),
	)
}

func TestFn(t *testing.T) {
	Test(t,
		That("fn f {|x| put x=$x'.' }; f lorem; f ipsum").
			Puts("x=lorem.", "x=ipsum."),
		// Recursive functions with fn. Regression test for #1206.
		That("fn f {|n| if (== $n 0) { num 1 } else { * $n (f (- $n 1)) } }; f 3").
			Puts(6),
		// Exception thrown by return is swallowed by a fn-defined function.
		That("fn f { put a; return; put b }; f").Puts("a"),

		// Error when evaluating the lambda
		That("fn f {|&opt=(fail x)| }").Throws(FailError{"x"}, "fail x"),
	)
}

// Regression test for #1225
func TestUse_SetsVariableCorrectlyIfModuleCallsExtendGlobal(t *testing.T) {
	libdir := InTempDir(t)

	ApplyDir(Dir{"a.elv": "add-var"})
	ev := NewEvaler()
	ev.LibDirs = []string{libdir}
	addVar := func() {
		ev.ExtendGlobal(BuildNs().AddVar("b", vars.NewReadOnly("foo")))
	}
	ev.ExtendBuiltin(BuildNs().AddGoFn("add-var", addVar))

	err := ev.Eval(parse.Source{Code: "use a"}, EvalCfg{})
	if err != nil {
		t.Fatal(err)
	}

	g := ev.Global()
	if g.IndexString("a:").Get().(*Ns) == nil {
		t.Errorf("$a: is nil")
	}
	if g.IndexString("b").Get().(string) != "foo" {
		t.Errorf(`$b is not "foo"`)
	}
}

func TestUse_SupportsCircularDependency(t *testing.T) {
	libdir := InTempDir(t)
	ApplyDir(Dir{
		"a.elv": "var pre = apre; use b; put $b:pre $b:post; var post = apost",
		"b.elv": "var pre = bpre; use a; put $a:pre $a:post; var post = bpost",
	})

	TestWithSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir} },
		That(`use a`).Puts(
			// When b.elv is imported from a.elv, $a:pre is set but $a:post is
			// not
			"apre", nil,
			// After a.elv imports b.elv, both $b:pre and $b:post are set
			"bpre", "bpost"),
	)
}

func TestUse(t *testing.T) {
	libdir1 := InTempDir(t)
	ApplyDir(Dir{
		"shadow.elv": "put lib1",
	})

	libdir2 := InTempDir(t)
	ApplyDir(Dir{
		"has-init.elv": "put has-init",
		"put-x.elv":    "put $x",
		"lorem.elv":    "name = lorem; fn put-name { put $name }",
		"d.elv":        "name = d",
		"shadow.elv":   "put lib2",
		"a": Dir{
			"b": Dir{
				"c": Dir{
					"d.elv": "name = a/b/c/d",
					"x.elv": "use ./d; d = $d:name; use ../../../lorem; lorem = $lorem:name",
				},
			},
		},
	})

	TestWithSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir1, libdir2} },
		That(`use lorem; put $lorem:name`).Puts("lorem"),
		// imports are lexically scoped
		// TODO: Support testing for compilation error
		That(`{ use lorem }; put $lorem:name`).DoesNotCompile(),

		// prefers lib dir that appear earlier
		That("use shadow").Puts("lib1"),

		// use of imported variable is captured in upvalue
		That(`use lorem; { put $lorem:name }`).Puts("lorem"),
		That(`{ use lorem; { put $lorem:name } }`).Puts("lorem"),
		That(`({ use lorem; put { { put $lorem:name } } })`).Puts("lorem"),
		// use of imported function is also captured in upvalue
		That(`{ use lorem; { lorem:put-name } }`).Puts("lorem"),

		// use of a nested module
		That(`use a/b/c/d; put $d:name`).Puts("a/b/c/d"),
		// module is cached after first use
		That(`use has-init; use has-init`).Puts("has-init"),
		// repeated uses result in the same namespace being imported
		That("use lorem; use lorem lorem2; put $lorem:name $lorem2:name").
			Puts("lorem", "lorem"),
		// overriding module
		That(`use d; put $d:name; use a/b/c/d; put $d:name`).
			Puts("d", "a/b/c/d"),
		// relative uses
		That(`use a/b/c/x; put $x:d $x:lorem`).Puts("a/b/c/d", "lorem"),
		// relative uses from top-level
		That(`use ./d; put $d:name`).Puts("d"),

		// Renaming module
		That(`use a/b/c/d mod; put $mod:name`).Puts("a/b/c/d"),

		// Variables defined in the default global scope is invisible from
		// modules
		That("x = foo; use put-x").Throws(AnyError),

		// Using an unknown module spec fails.
		That("use unknown").Throws(ErrorWithType(NoSuchModule{})),
		That("use ./unknown").Throws(ErrorWithType(NoSuchModule{})),
		That("use ../unknown").Throws(ErrorWithType(NoSuchModule{})),

		// Nonexistent module
		That("use non-existent").Throws(ErrorWithMessage("no such module: non-existent")),

		// Wrong uses of "use".
		That("use").DoesNotCompile(),
		That("use a b c").DoesNotCompile(),
	)
}

// Regression test for #1072
func TestUse_WarnsAboutDeprecatedFeatures(t *testing.T) {
	progtest.SetDeprecationLevel(t, 17)
	libdir := InTempDir(t)
	MustWriteFile("dep.elv", "fn x { dir-history }")

	TestWithSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir} },
		// Importing module triggers check for deprecated features
		That("use dep").PrintsStderrWith("is deprecated"),
	)
}

package eval_test

import (
	"errors"
	"path/filepath"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/testutil"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestPragma(t *testing.T) {
	Test(t,
		That("pragma unknown-command").DoesNotCompile("need literal ="),
		That("pragma unknown-command =").DoesNotCompile("need pragma value"),
		That("pragma unknown-command x").DoesNotCompile("must be literal ="),
		That("pragma bad-name = some-value").DoesNotCompile("unknown pragma bad-name"),
		That("pragma unknown-command = bad").DoesNotCompile("invalid value for unknown-command: bad"),
	)
	// Actual effect of the unknown-command pragma is tested in TestCommand_External
}

func TestVar(t *testing.T) {
	// NOTE: TestClosure has more tests for the interaction between assignment
	// and variable scoping.

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
		That("var cmd~; is $cmd~ $nop~").Puts(true),
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

		// Concurrently creating a new variable and accessing existing variable.
		// Run with "go test -race".
		That("var x = 1", "put $x | var y = (all)").DoesNothing(),
		That("nop (var x = 1) | nop").DoesNothing(),

		// Assignment errors when the RHS errors.
		That("var x = [][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
		// Arity mismatch.
		That("var x = 1 2").Throws(
			errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"var x = 1 2"),
		That("var x y = 1").Throws(
			errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: 2, ValidHigh: 2, Actual: 1},
			"var x y = 1"),
		That("var x y @z = 1").Throws(
			errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: 2, ValidHigh: -1, Actual: 1},
			"var x y @z = 1"),

		// Variable name must not be empty
		That("var ''").DoesNotCompile("variable name must not be empty"),
		// Variable name that must be quoted after $ must be quoted
		That("var a/b").DoesNotCompile("lvalue must be valid literal variable names"),
		// Multiple @ not allowed
		That("var x @y @z = a b c d").DoesNotCompile("at most one rest variable is allowed"),
		// Non-local not allowed
		That("var ns:a").DoesNotCompile("cannot create variable $ns:a; new variables can only be created in the current scope"),
		// Index not allowed
		That("var a[0]").DoesNotCompile("new variable $a must not have indices"),
		// Composite expression not allowed
		That("var a'b'").DoesNotCompile("lvalue may not be composite expressions"),
		// Braced lists must not have any indices when used as a lvalue.
		That("var {a b}[0] = x y").DoesNotCompile("braced list may not have indices when used as lvalue"),
	)
}

func TestSet(t *testing.T) {
	Test(t,
		// Setting one variable
		That("var x; set x = foo", "put $x").Puts("foo"),
		// An empty RHS is technically legal although rarely useful.
		That("var x; set @x =", "put $x").Puts(vals.EmptyList),
		// Variable must already exist
		That("set x = foo").DoesNotCompile("cannot find variable $x"),
		// List element assignment
		That("var li = [foo bar]; set li[0] = 233; put $@li").Puts("233", "bar"),
		// Variable in list assignment must already be defined. Regression test
		// for b.elv.sh/889.
		That("set y[0] = a").DoesNotCompile("cannot find variable $y"),
		// Map element assignment
		That("var di = [&k=v]; set di[k] = lorem; set di[k2] = ipsum",
			"put $di[k] $di[k2]").Puts("lorem", "ipsum"),
		That("var d = [&a=[&b=v]]; put $d[a][b]; set d[a][b] = u; put $d[a][b]").
			Puts("v", "u"),
		That("var li = [foo]; set li[(fail foo)] = bar").Throws(FailError{"foo"}),
		That("var li = [foo]; set li[0 1] = foo bar").
			Throws(ErrorWithMessage("multi indexing not implemented")),
		That("var li = [[]]; set li[1][2] = bar").
			Throws(errs.OutOfRange{What: "index",
				ValidLow: "0", ValidHigh: "0", Actual: "1"}, "li[1][2]"),

		// Assignment to read-only var is a compile-time error.
		That("set nil = 1").DoesNotCompile("variable $nil is read-only"),
		That("var a b; set a true b = 1 2 3").DoesNotCompile("variable $true is read-only"),
		That("set @true = 1").DoesNotCompile("variable $true is read-only"),
		That("var r; set true @r = 1").DoesNotCompile("variable $true is read-only"),
		That("var r; set @r true = 1").DoesNotCompile("variable $true is read-only"),

		// Error conditions already covered by TestVar are not repeated.

		// = is required.
		That("var x; set x").DoesNotCompile("need = and right-hand-side"),

		// set a non-exist environment
		That("has-env X; set E:X = x; get-env X; unset-env X").
			Puts(false, "x"),
	)
}

func TestSet_ErrorInSetMethod(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *Evaler) { addBadVar(ev, 0) },
		That("set bad = foo").Throws(errBadVar, "bad"),
		That("var a; set bad @a = foo").Throws(errBadVar, "bad"),
		That("var a; set a @bad = foo").Throws(errBadVar, "@bad"),
		That("var a; set @a bad = foo").Throws(errBadVar, "bad"),
	)
}

func TestTmp(t *testing.T) {
	testutil.Unsetenv(t, "X")

	Test(t,
		That("var x = foo; put $x; { tmp x = bar; put $x }; put $x").
			Puts("foo", "bar", "foo"),

		That("var x; tmp x = y").DoesNotCompile("tmp may only be used inside a function"),
		That("{ tmp x = y }").DoesNotCompile("cannot find variable $x"),

		That("has-env X; { tmp E:X = y; put $E:X }; has-env X; put $E:X").
			Puts(false, "y", false, ""),
		That("set-env X x; { tmp E:X = y; put $E:X }; get-env X; put $E:X; unset-env X").
			Puts("y", "x", "x"),
	)
}

func TestTmp_ErrorSetting(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *Evaler) { addBadVar(ev, 0) },
		That("{ tmp bad = foo }").Throws(errBadVar, "bad", "{ tmp bad = foo }"),
	)
}

func TestTmp_ErrorRestoring(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *Evaler) { addBadVar(ev, 1) },
		That("{ tmp bad = foo; put after }").
			Puts("after").
			Throws(ErrorWithMessage("restore variable: bad var"),
				"bad", "{ tmp bad = foo; put after }"),
	)
}

func addBadVar(ev *Evaler, allowedSets int) {
	ev.ExtendGlobal(BuildNs().AddVar("bad", &badVar{allowedSets}))
}

var errBadVar = errors.New("bad var")

type badVar struct{ allowedSets int }

func (v *badVar) Get() any { return nil }

func (v *badVar) Set(any) error {
	if v.allowedSets == 0 {
		return errBadVar
	}
	v.allowedSets--
	return nil
}

func TestDel(t *testing.T) {
	testutil.Setenv(t, "TEST_ENV", "test value")

	Test(t,
		// Deleting variable
		That("var x = 1; del x").DoesNothing(),
		That("var x = 1; del x; echo $x").DoesNotCompile("variable $x not found"),
		// Deleting environment variable
		That("has-env TEST_ENV", "del E:TEST_ENV", "has-env TEST_ENV").Puts(true, false),
		// Deleting variable whose name contains special characters
		That("var 'a/b' = foo; del 'a/b'").DoesNothing(),
		// Deleting element
		That("var x = [&k=v &k2=v2]; del x[k2]; keys $x").Puts("k"),
		That("var x = [[&k=v &k2=v2]]; del x[0][k2]; keys $x[0]").Puts("k"),

		// Error cases

		// Deleting nonexistent variable
		That("del x").DoesNotCompile("no variable $x"),
		// Deleting element of nonexistent variable
		That("del x[0]").DoesNotCompile("no variable $x"),
		// Deleting variable in non-local namespace
		That("var a: = (ns [&b=$nil])", "del a:b").DoesNotCompile("only variables in the local scope or E: can be deleted"),
		// Variable name given with $
		That("var x = 1; del $x").DoesNotCompile("arguments to del must omit the dollar sign"),
		// Variable name not given as a single primary expression
		That("var ab = 1; del a'b'").DoesNotCompile("arguments to del must be variable or variable elements"),
		// Variable name not a string
		That("del [a]").DoesNotCompile("arguments to del must be variable or variable elements"),
		// Variable name has sigil
		That("var x = []; del @x").DoesNotCompile("arguments to del must be variable or variable elements"),
		// Variable name not quoted when it should be
		That("var 'a/b' = foo; del a/b").DoesNotCompile("arguments to del must be variable or variable elements"),

		// Index is multiple values
		That("var x = [&k1=v1 &k2=v2]", "del x[k1 k2]").Throws(
			ErrorWithMessage("index must evaluate to a single value in argument to del"),
			"k1 k2"),
		// Index expression throws exception
		That("var x = [&k]", "del x[(fail x)]").Throws(FailError{"x"}, "fail x"),
		// Value does not support element removal
		That("var x = (num 1)", "del x[k]").Throws(
			ErrorWithMessage("value does not support element removal"),
			// TODO: Fix the stack trace so that it is "x[k]"
			"x[k"),
		// Intermediate element does not exist
		That("var x = [&]", "del x[k][0]").Throws(
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
		That("var x = a; and $false (x = b); put $x").Puts(false, "a"),

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
		That("var x = a; or $true (x = b); put $x").Puts(true, "a"),

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

func TestSpecialFormThunks(t *testing.T) {
	// Regression test for b.elv.sh/1456
	Test(t,
		That("for x [] {|arg| }").DoesNotCompile("for body must not have arguments"),
		That("for x [] {|&opt=val| }").DoesNotCompile("for body must not have options"),
		// The other special forms use the same utility under the hood and are
		// not repeated
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
		That("try { nop } catch { put bad } else { put good }").Puts("good"),
		That("try { fail tr } catch - { put bad } else { put good }").
			Puts("bad"),
		That("try { fail tr } finally { put final }").
			Puts("final").
			Throws(ErrorWithMessage("tr")),

		That("try { fail tr } catch { fail ex } finally { put final }").
			Puts("final").
			Throws(ErrorWithMessage("ex")),

		That("try { fail tr } catch { put ex } finally { fail final }").
			Puts("ex").
			Throws(ErrorWithMessage("final")),

		That("try { fail tr } catch { fail ex } finally { fail final }").
			Throws(ErrorWithMessage("final")),

		// Must have catch or finally
		That("try { fail tr }").DoesNotCompile("try must be followed by a catch block or a finally block"),
		// Rest variable not allowed
		That("try { nop } catch @a { }").DoesNotCompile("rest variable not allowed"),

		// A readonly var as a target for the "catch" clause is a compile-time
		// error.
		That("try { fail reason } catch nil { }").DoesNotCompile("variable $nil is read-only"),
		That("try { fail reason } catch x { }").DoesNothing(),

		// A quoted var name, that would be invalid as a bareword, should be allowed as the referent
		// in a `try...catch...` block.
		That("try { fail hard } catch 'x=' { put $'x='[reason][type] }").Puts("fail"),

		// Regression test: "try { } catch" is a syntax error, but it should not
		// panic.
		That("try { } catch").DoesNotCompile("need variable or body"),
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
		That("var x = 0; while (< $x 4) { fail haha }").Throws(FailError{"haha"}),
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
		That("for {x,y} [] { }").DoesNotCompile("must be exactly one lvalue"),
		// Invalid for loop lvalue. You can't use a var in a namespace other
		// than the local namespace as the lvalue in a for loop.
		That("for no-such-namespace:x [a b] { }").DoesNotCompile("cannot create variable $no-such-namespace:x; new variables can only be created in the current scope"),
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
	libdir := testutil.InTempDir(t)

	testutil.ApplyDir(testutil.Dir{"a.elv": "add-var"})
	ev := NewEvaler()
	ev.LibDirs = []string{libdir}
	addVar := func() {
		ev.ExtendGlobal(BuildNs().AddVar("b", vars.NewReadOnly("foo")))
	}
	ev.ExtendBuiltin(BuildNs().AddGoFn("add-var", addVar))

	err := ev.Eval(parse.Source{Name: "[test]", Code: "use a"}, EvalCfg{})
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
	libdir := testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"a.elv": "var pre = apre; use b; put $b:pre $b:post; var post = apost",
		"b.elv": "var pre = bpre; use a; put $a:pre $a:post; var post = bpost",
	})

	TestWithEvalerSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir} },
		That(`use a`).Puts(
			// When b.elv is imported from a.elv, $a:pre is set but $a:post is
			// not
			"apre", nil,
			// After a.elv imports b.elv, both $b:pre and $b:post are set
			"bpre", "bpost"),
	)
}

func TestUse(t *testing.T) {
	libdir1 := testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"shadow.elv":       "put lib1",
		"invalid-utf8.elv": "\xff",
	})

	libdir2 := testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"has-init.elv": "put has-init",
		"put-x.elv":    "put $x",
		"lorem.elv":    "var name = lorem; fn put-name { put $name }",
		"d.elv":        "var name = d",
		"shadow.elv":   "put lib2",
		"a": testutil.Dir{
			"b": testutil.Dir{
				"c": testutil.Dir{
					"d.elv": "var name = a/b/c/d",
					"x.elv": "use ./d; var d = $d:name; use ../../../lorem; var lorem = $lorem:name",
				},
			},
		},
	})

	TestWithEvalerSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir1, libdir2} },
		That(`use lorem; put $lorem:name`).Puts("lorem"),
		// imports are lexically scoped
		// TODO: Support testing for compilation error
		That(`{ use lorem }; put $lorem:name`).DoesNotCompile("variable $lorem:name not found"),

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
		That("var x = foo; use put-x").Throws(ErrorWithType(&CompilationError{})),

		// Using an unknown module spec fails.
		That("use unknown").Throws(ErrorWithType(NoSuchModule{})),
		That("use ./unknown").Throws(ErrorWithType(NoSuchModule{})),
		That("use ../unknown").Throws(ErrorWithType(NoSuchModule{})),

		// Invalid UTF-8 in module file
		That("use invalid-utf8").Throws(ErrorWithMessage(
			filepath.Join(libdir1, "invalid-utf8.elv")+": source is not valid UTF-8")),

		// Nonexistent module
		That("use non-existent").Throws(ErrorWithMessage("no such module: non-existent")),

		// Wrong uses of "use".
		That("use").DoesNotCompile("need module spec"),
		That("use a b c").DoesNotCompile("superfluous arguments"),
	)
}

// Regression test for #1072
func TestUse_WarnsAboutDeprecatedFeatures(t *testing.T) {
	testutil.Set(t, &prog.DeprecationLevel, 18)
	libdir := testutil.InTempDir(t)
	must.WriteFile("dep.elv", "a=b nop $a")

	TestWithEvalerSetup(t, func(ev *Evaler) { ev.LibDirs = []string{libdir} },
		// Importing module triggers check for deprecated features
		That("use dep").PrintsStderrWith("is deprecated"),
	)
}

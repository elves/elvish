package eval_test

import (
	"path/filepath"
	"strings"
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/errs"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/testutil"
)

func TestBuiltinSpecial(t *testing.T) {
	Test(t,
		// del - deleting variable
		That("x = 1; del x").DoesNothing(),
		That("x = 1; del x; echo $x").DoesNotCompile(),
		That("x = 1; del :x; echo $x").DoesNotCompile(),
		That("x = 1; del local:x; echo $x").DoesNotCompile(),
		// del - deleting element
		That("x = [&k=v &k2=v2]; del x[k2]; keys $x").Puts("k"),
		That("x = [[&k=v &k2=v2]]; del x[0][k2]; keys $x[0]").Puts("k"),
		// del - wrong use of del
		That("del x").DoesNotCompile(),
		That("del x[0]").DoesNotCompile(),
		That("ab = 1; del a'b'").DoesNotCompile(),
		That("del a:b").DoesNotCompile(),
		That("x = 1; del $x").DoesNotCompile(),
		That("del [a]").DoesNotCompile(),
		That("x = []; del @x").DoesNotCompile(),

		// and
		That("and $true $false").Puts(false),
		// and - short circuit
		That("x = a; and $false (x = b); put $x").Puts(false, "a"),

		// or
		That("or $true $false").Puts(true),
		That("or a b").Puts("a"),
		// or - short circuit
		That("x = a; or $true (x = b); put $x").Puts(true, "a"),

		// if
		That("if true { put then }").Puts("then"),
		That("if $false { put then } else { put else }").Puts("else"),
		That("if $false { put 1 } elif $false { put 2 } else { put 3 }").
			Puts("3"),
		That("if $false { put 2 } elif true { put 2 } else { put 3 }").Puts("2"),

		// try
		That("try { nop } except { put bad } else { put good }").Puts("good"),
		That("try { e:false } except - { put bad } else { put good }").
			Puts("bad"),
		That("try { fail tr }").Throws(ErrorWithMessage("tr")),
		That("try { fail tr } finally { put final }").
			Puts("final").Throws(ErrorWithMessage(
			"tr")),

		That("try { fail tr } except { fail ex } finally { put final }").
			Puts("final").Throws(ErrorWithMessage(
			"ex")),

		That("try { fail tr } except { put ex } finally { fail final }").
			Puts("ex").Throws(ErrorWithMessage(
			"final")),

		That("try { fail tr } except { fail ex } finally { fail final }").Throws(ErrorWithMessage(
			"final")),

		// try - wrong use
		That("try { nop } except @a { }").DoesNotCompile(),

		// while
		That("x=0; while (< $x 4) { put $x; x=(+ $x 1) }").
			Puts("0", 1.0, 2.0, 3.0),
		That("x = 0; while (< $x 4) { put $x; break }").Puts("0"),
		That("x = 0; while (< $x 4) { fail haha }").Throws(AnyError),
		That("x = 0; while (< $x 4) { put $x; x=(+ $x 1) } else { put bad }").
			Puts("0", 1.0, 2.0, 3.0),
		That("while $false { put bad } else { put good }").Puts("good"),

		// for
		That("for x [tempora mores] { put 'O '$x }").
			Puts("O tempora", "O mores"),
		// break
		That("for x [a] { break } else { put $x }").DoesNothing(),
		// else
		That("for x [a] { put $x } else { put $x }").Puts("a"),
		// continue
		That("for x [a b] { put $x; continue; put $x; }").Puts("a", "b"),
		// More than one iterator.
		That("for {x,y} [] { }").DoesNotCompile(),
		// Invalid for loop lvalue. You can't use a var in a namespace other
		// than the local namespace as the lvalue in a for loop.
		That("for no-such-namespace:x [a b] { }").DoesNotCompile(),
		// Exception when evaluating iterable.
		That("for x [][0] { }").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		// More than one iterable.
		That("for x (put a b) { }").Throws(
			errs.ArityMismatch{
				What:     "value being iterated",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"(put a b)"),

		// fn.
		That("fn f [x]{ put x=$x'.' }; f lorem; f ipsum").
			Puts("x=lorem.", "x=ipsum."),
		// Recursive functions with fn. Regression test for #1206.
		That("fn f [n]{ if (== $n 0) { put 1 } else { * $n (f (- $n 1)) } }; f 3").
			Puts(6.0),

		// return.
		That("fn f []{ put a; return; put b }; f").Puts("a"),
	)
}

func TestUse(t *testing.T) {
	libdir, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustMkdirAll(filepath.Join("a", "b", "c"))

	writeMod := func(name, content string) {
		fname := filepath.Join(strings.Split(name, "/")...) + ".elv"
		testutil.MustWriteFile(fname, []byte(content), 0600)
	}
	writeMod("has-init", "put has-init")
	writeMod("put-x", "put $x")
	writeMod("lorem", "name = lorem; fn put-name { put $name }")
	writeMod("d", "name = d")
	writeMod("a/b/c/d", "name = a/b/c/d")
	writeMod("a/b/c/x",
		"use ./d; d = $d:name; use ../../../lorem; lorem = $lorem:name")

	TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) },
		That(`use lorem; put $lorem:name`).Puts("lorem"),
		// imports are lexically scoped
		// TODO: Support testing for compilation error
		// That(`{ use lorem }; put $lorem:name`).ErrorsAny(),

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

		// TODO: Test module namespace

		// Wrong uses of "use".
		That("use").DoesNotCompile(),
		That("use a b c").DoesNotCompile(),
	)
}

// Regression test for #1072
func TestUse_WarnsAboutDeprecatedFeatures(t *testing.T) {
	restore := prog.SetShowDeprecations(true)
	defer restore()
	libdir, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.MustWriteFile("dep.elv", []byte("x = (ord 1)"), 0600)

	TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) },
		// Importing module triggers check for deprecated features
		That("use dep").PrintsStderrWith("is deprecated"),
	)
}

package eval

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestBuiltinSpecial(t *testing.T) {
	Test(t,
		// del - deleting variable
		That("x = 1; del $x; echo $x").DoesNotCompile(),
		That("x = 1; del $:x; echo $x").DoesNotCompile(),
		That("x = 1; del $local:x; echo $x").DoesNotCompile(),
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
		That("try { fail tr }").Throws(errors.New("tr")),
		That("try { fail tr } finally { put final }").
			Puts("final").Throws(errors.New("tr")),
		That("try { fail tr } except { fail ex } finally { put final }").
			Puts("final").Throws(errors.New("ex")),
		That("try { fail tr } except { put ex } finally { fail final }").
			Puts("ex").Throws(errors.New("final")),
		That("try { fail tr } except { fail ex } finally { fail final }").
			Throws(errors.New("final")),
		// try - wrong use
		That("try { nop } except @a { }").DoesNotCompile(),

		// while
		That("x=0; while (< $x 4) { put $x; x=(+ $x 1) }").
			Puts("0", 1.0, 2.0, 3.0),
		That("x = 0; while (< $x 4) { put $x; break }").Puts("0"),
		That("x = 0; while (< $x 4) { fail haha }").ThrowsAny(),
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

		// fn.
		That("fn f [x]{ put x=$x'.' }; f lorem; f ipsum").
			Puts("x=lorem.", "x=ipsum."),
		// fn with good and bad function names. The first validates that
		// non-ASCII chars are handled correctly.
		That("fn blåbær [x]{ put $x }; blåbær argle").Puts("argle"),
		That("fn f.bad [x]{ put $x }").DoesNotCompile(),
		// return.
		That("fn f []{ put a; return; put b }; f").Puts("a"),
	)
}

var useTests = []TestCase{}

func TestUse(t *testing.T) {
	libdir, cleanup := util.InTestDir()
	defer cleanup()

	MustMkdirAll(filepath.Join("a", "b", "c"), 0700)

	writeMod := func(name, content string) {
		fname := filepath.Join(strings.Split(name, "/")...) + ".elv"
		MustWriteFile(fname, []byte(content), 0600)
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
		That("x = foo; use put-x").ThrowsAny(),

		// TODO: Test module namespace

		// Wrong uses of "use".
		That("use").DoesNotCompile(),
		That("use a b c").DoesNotCompile(),
		That("use a.c").DoesNotCompile(),
		That("use lib.dir/os dir/os").DoesNotCompile(),
		// Note that the period in the path is okay so this should compile but
		// fail because the module doesn't exist.
		That("use lib.dir/os os").ThrowsAny(),
	)
}

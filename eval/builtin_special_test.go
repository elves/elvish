package eval

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elves/elvish/util"
)

var builtinSpecialTests = []TestCase{
	// del
	That("x = [&k=v &k2=v2]; del x[k2]; keys $x").Puts("k"),
	That("x = [[&k=v &k2=v2]]; del x[0][k2]; keys $x[0]").Puts("k"),

	// if
	That("if true { put then }").Puts("then"),
	That("if $false { put then } else { put else }").Puts("else"),
	That("if $false { put 1 } elif $false { put 2 } else { put 3 }").Puts("3"),
	That("if $false { put 2 } elif true { put 2 } else { put 3 }").Puts("2"),

	// try
	That("try { nop } except { put bad } else { put good }").Puts("good"),
	That("try { e:false } except - { put bad } else { put good }").Puts("bad"),
	That("try { fail tr }").ErrorsWith(errors.New("tr")),
	That("try { fail tr } finally { put final }").Puts(
		"final").ErrorsWith(errors.New("tr")),
	That("try { fail tr } except { fail ex } finally { put final }").Puts(
		"final").ErrorsWith(errors.New("ex")),
	That("try { fail tr } except { put ex } finally { fail final }").Puts(
		"ex").ErrorsWith(errors.New("final")),
	That("try { fail tr } except { fail ex } finally { fail final }").ErrorsWith(
		errors.New("final")),

	// while
	That("x=0; while (< $x 4) { put $x; x=(+ $x 1) }").Puts("0", "1", "2", "3"),
	That("x = 0; while (< $x 4) { put $x; break }").Puts("0"),
	That("x = 0; while (< $x 4) { fail haha }").Errors(),

	// for
	That("for x [tempora mores] { put 'O '$x }").Puts("O tempora", "O mores"),
	// break
	That("for x [a] { break } else { put $x }").DoesNothing(),
	// else
	That("for x [a] { put $x } else { put $x }").Puts("a"),
	// continue
	That("for x [a b] { put $x; continue; put $x; }").Puts("a", "b"),

	// fn.
	That("fn f [x]{ put x=$x'.' }; f lorem; f ipsum").Puts(
		"x=lorem.", "x=ipsum."),
	// return.
	That("fn f []{ put a; return; put b }; f").Puts("a"),
}

func TestBuiltinSpecial(t *testing.T) {
	Test(t, builtinSpecialTests)
}

var useTests = []TestCase{
	That(`use lorem; put $lorem:name`).Puts("lorem"),
	// imports are lexically scoped
	// TODO: Support testing for compilation error
	// That(`{ use lorem }; put $lorem:name`).ErrorsAny(),

	// use of imported variable is captured in upvalue
	That(`({ use lorem; put { { put $lorem:name } } })`).Puts("lorem"),
	// use of imported function is also captured in upvalue
	That(`{ use lorem; { lorem:put-name } }`).Puts("lorem"),

	// use of a nested module
	That(`use a:b/c/d; put $d:name`).Puts("a/b/c/d"),
	// module is cached after first use
	That(`use has-init; use has-init`).Puts("has-init"),
	// overriding module
	That(`use d; put $d:name; use a/b/c/d; put $d:name`).Puts("d", "a/b/c/d"),
	// relative uses
	That(`use a/b/c/x; put $x:d $x:lorem`).Puts("a/b/c/d", "lorem"),

	// Variables defined in the default global scope is invisible from modules
	That("x = foo; use put-x").Errors(),

	// TODO: Test module namespace
}

func TestUse(t *testing.T) {
	util.InTempDir(func(libdir string) {
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

		TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) }, useTests)
	})
}

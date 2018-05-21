package eval

import (
	"testing"

	"github.com/elves/elvish/util"
)

func TestBuiltinFnMisc(t *testing.T) {
	Test(t, []TestCase{
		That(`f = (mktemp elvXXXXXX); echo 'put x' > $f
		      -source $f; rm $f`).Puts("x"),
		That(`f = (mktemp elvXXXXXX); echo 'put $x' > $f
		      fn p [x]{ -source $f }; p x; rm $f`).Puts("x"),
	})
}

func TestResolve(t *testing.T) {
	util.InTempDir(func(libdir string) {
		MustWriteFile("mod.elv", []byte("fn func { }"), 0600)

		TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) }, []TestCase{
			That("resolve for").Puts("special"),
			That("resolve put").Puts("$put~"),
			That("fn f { }; resolve f").Puts("$f~"),
			That("use mod; resolve mod:func").Puts("$mod:func~"),
			That("resolve cat").Puts("(external cat)"),
		})
	})
}

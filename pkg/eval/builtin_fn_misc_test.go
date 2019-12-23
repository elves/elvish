package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestBuiltinFnMisc(t *testing.T) {
	Test(t,
		That(`f = (constantly foo); $f; $f`).Puts("foo", "foo"),

		That(`f = (mktemp elvXXXXXX); echo 'put x' > $f
		      -source $f; rm $f`).Puts("x"),
		That(`f = (mktemp elvXXXXXX); echo 'put $x' > $f
		      fn p [x]{ -source $f }; p x; rm $f`).Puts("x"),
	)
}

func TestResolve(t *testing.T) {
	libdir, cleanup := util.InTestDir()
	defer cleanup()

	MustWriteFile("mod.elv", []byte("fn func { }"), 0600)

	TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) },
		That("resolve for").Puts("special"),
		That("resolve put").Puts("$put~"),
		That("fn f { }; resolve f").Puts("$f~"),
		That("use mod; resolve mod:func").Puts("$mod:func~"),
		That("resolve cat").Puts("(external cat)"),
	)
}

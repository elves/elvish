package eval

import (
	"errors"
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

		// Test the "time" builtin.
		//
		// Since runtime duration is non-deterministic, we only have some sanity
		// checks here.
		That("time { echo foo } | a _ = (all)", "put $a").Puts("foo"),
		That("duration = ''",
			"time &on-end=[x]{ duration = $x } { echo foo } | out = (all)",
			"put $out", "kind-of $duration").Puts("foo", "number"),
		That("time { fail body } | nop (all)").ThrowsCause(errors.New("body")),
		That("time &on-end=[_]{ fail on-end } { }").
			ThrowsCause(errors.New("on-end")),
		That("time &on-end=[_]{ fail on-end } { fail body }").
			ThrowsCause(errors.New("body")),
	)
}

func TestResolve(t *testing.T) {
	libdir, cleanup := util.InTestDir()
	defer cleanup()

	mustWriteFile("mod.elv", []byte("fn func { }"), 0600)

	TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) },
		That("resolve for").Puts("special"),
		That("resolve put").Puts("$put~"),
		That("fn f { }; resolve f").Puts("$f~"),
		That("use mod; resolve mod:func").Puts("$mod:func~"),
		That("resolve cat").Puts("(external cat)"),
	)
}

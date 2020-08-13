package eval

import (
	"errors"
	"testing"
	"time"

	"github.com/elves/elvish/pkg/util"
)

func TestBuiltinFnMisc(t *testing.T) {
	Test(t,
		That(`f = (constantly foo); $f; $f`).Puts("foo", "foo"),

		That("eval 'put x'").Puts("x"),
		// Using initial binding in &ns.
		That("n = (ns [&x=foo]); eval 'put $x' &ns=$n").Puts("foo"),
		// Altering variables in &ns.
		That("n = (ns [&x=foo]); eval 'x = bar' &ns=$n; put $n[x]").Puts("bar"),
		// Parse error.
		That("eval '['").ThrowsAny(),
		// Compilation error.
		That("eval 'put $x'").ThrowsAny(),
		// Exception.
		That("eval 'fail x'").ThrowsCause(FailError{"x"}),

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
		That("time { fail body } | nop (all)").ThrowsCause(FailError{"body"}),
		That("time &on-end=[_]{ fail on-end } { }").
			ThrowsCause(FailError{"on-end"}),
		That("time &on-end=[_]{ fail on-end } { fail body }").
			ThrowsCause(FailError{"body"}),
	)
}

func TestUseMod(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()
	mustWriteFile("mod.elv", []byte("x = value"), 0600)

	Test(t,
		That("put (use-mod ./mod)[x]").Puts("value"),
	)
}

func timeAfterMock(fm *Frame, d time.Duration) <-chan time.Time {
	fm.ports[1].Chan <- d // report to the test framework the duration we received
	return time.After(0)
}

func TestSleep(t *testing.T) {
	timeAfter = timeAfterMock
	Test(t,
		That(`sleep 0`).Puts(0*time.Second),
		That(`sleep 1`).Puts(1*time.Second),
		That(`sleep 1.3s`).Puts(1300*time.Millisecond),
		That(`sleep 0.1`).Puts(100*time.Millisecond),
		That(`sleep 0.1ms`).Puts(100*time.Microsecond),
		That(`sleep 3h5m7s`).Puts((3*3600+5*60+7)*time.Second),

		That(`sleep 1x`).Throws(errors.New("invalid sleep duration"), "sleep 1x"),
		That(`sleep -7`).Throws(errors.New("sleep duration must be >= zero"), "sleep -7"),
		That(`sleep -3h`).Throws(errors.New("sleep duration must be >= zero"), "sleep -3h"),

		// Verify the correct behavior if a number rather than a string is
		// passed to the command.
		That(`sleep (float64 0)`).Puts(0*time.Second),
		That(`sleep (float64 1.7)`).Puts(1700*time.Millisecond),
		That(`sleep (float64 -7)`).Throws(
			errors.New("sleep duration must be >= zero"),
			"sleep (float64 -7)"),

		// An invalid argument type should raise an exception.
		That(`sleep [1]`).Throws(errors.New("invalid sleep duration"), "sleep [1]"),
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
		That(`resolve external`).Puts("$external~"),
	)
}

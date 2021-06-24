package eval_test

import (
	"os"
	"testing"
	"time"

	. "src.elv.sh/pkg/eval"

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
		That(`f = (constantly foo); $f; $f`).Puts("foo", "foo"),
		thatOutputErrorIsBubbled("(constantly foo)"),
	)
}

func TestEval(t *testing.T) {
	Test(t,
		That("eval 'put x'").Puts("x"),
		// Using variable from the local scope.
		That("x = foo; eval 'put $x'").Puts("foo"),
		// Setting a variable in the local scope.
		That("x = foo; eval 'x = bar'; put $x").Puts("bar"),
		// Using variable from the upvalue scope.
		That("x = foo; { nop $x; eval 'put $x' }").Puts("foo"),
		// Specifying a namespace.
		That("n = (ns [&x=foo]); eval 'put $x' &ns=$n").Puts("foo"),
		// Altering variables in the specified namespace.
		That("n = (ns [&x=foo]); eval 'x = bar' &ns=$n; put $n[x]").Puts("bar"),
		// Newly created variables do not appear in the local namespace.
		That("eval 'x = foo'; put $x").DoesNotCompile(),
		// Newly created variables do not alter the specified namespace, either.
		That("n = (ns [&]); eval &ns=$n 'x = foo'; put $n[x]").
			Throws(vals.NoSuchKey("x"), "$n[x]"),
		// However, newly created variable can be accessed in the final
		// namespace using &on-end.
		That("eval &on-end=[n]{ put $n[x] } 'x = foo'").Puts("foo"),
		// Parse error.
		That("eval '['").Throws(AnyError),
		// Compilation error.
		That("eval 'put $x'").Throws(AnyError),
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

func TestTime(t *testing.T) {
	Test(t,
		// Since runtime duration is non-deterministic, we only have some sanity
		// checks here.
		That("time { echo foo } | a _ = (all)", "put $a").Puts("foo"),
		That("duration = ''",
			"time &on-end=[x]{ duration = $x } { echo foo } | out = (all)",
			"put $out", "kind-of $duration").Puts("foo", "number"),
		That("time { fail body } | nop (all)").Throws(FailError{"body"}),
		That("time &on-end=[_]{ fail on-end } { }").Throws(
			FailError{"on-end"}),

		That("time &on-end=[_]{ fail on-end } { fail body }").Throws(
			FailError{"body"}),

		thatOutputErrorIsBubbled("time { }"),
	)
}

func TestUseMod(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.MustWriteFile("mod.elv", []byte("x = value"), 0600)

	Test(t,
		That("put (use-mod ./mod)[x]").Puts("value"),
	)
}

func timeAfterMock(fm *Frame, d time.Duration) <-chan time.Time {
	fm.ValueOutput().Put(d) // report to the test framework the duration we received
	return time.After(0)
}

func TestSleep(t *testing.T) {
	TimeAfter = timeAfterMock
	Test(t,
		That(`sleep 0`).Puts(0*time.Second),
		That(`sleep 1`).Puts(1*time.Second),
		That(`sleep 1.3s`).Puts(1300*time.Millisecond),
		That(`sleep 0.1`).Puts(100*time.Millisecond),
		That(`sleep 0.1ms`).Puts(100*time.Microsecond),
		That(`sleep 3h5m7s`).Puts((3*3600+5*60+7)*time.Second),

		That(`sleep 1x`).Throws(ErrInvalidSleepDuration, "sleep 1x"),
		That(`sleep -7`).Throws(ErrNegativeSleepDuration, "sleep -7"),
		That(`sleep -3h`).Throws(ErrNegativeSleepDuration, "sleep -3h"),

		That(`sleep 1/2`).Puts(time.Second/2), // rational number string

		// Verify the correct behavior if a numeric type, rather than a string, is passed to the
		// command.
		That(`sleep (num 42)`).Puts(42*time.Second),
		That(`sleep (float64 0)`).Puts(0*time.Second),
		That(`sleep (float64 1.7)`).Puts(1700*time.Millisecond),
		That(`sleep (float64 -7)`).Throws(ErrNegativeSleepDuration, "sleep (float64 -7)"),

		// An invalid argument type should raise an exception.
		That(`sleep [1]`).Throws(ErrInvalidSleepDuration, "sleep [1]"),
	)
}

func TestResolve(t *testing.T) {
	libdir, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustWriteFile("mod.elv", []byte("fn func { }"), 0600)

	TestWithSetup(t, func(ev *Evaler) { ev.SetLibDir(libdir) },
		That("resolve for").Puts("special"),
		That("resolve put").Puts("$put~"),
		That("fn f { }; resolve f").Puts("$f~"),
		That("use mod; resolve mod:func").Puts("$mod:func~"),
		That("resolve cat").Puts("(external cat)"),
		That(`resolve external`).Puts("$external~"),
	)
}

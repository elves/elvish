package eval_test

import (
	"os"
	"testing"
	"time"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestSleep(t *testing.T) {
	testutil.Set(t, TimeAfter,
		func(fm *Frame, d time.Duration) <-chan time.Time {
			fm.ValueOutput().Put(d)
			return time.After(0)
		})

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
		That(`sleep (num 0)`).Puts(0*time.Second),
		That(`sleep (num 1.7)`).Puts(1700*time.Millisecond),
		That(`sleep (num -7)`).Throws(ErrNegativeSleepDuration, "sleep (num -7)"),

		// An invalid argument type should raise an exception.
		That(`sleep [1]`).Throws(ErrInvalidSleepDuration, "sleep [1]"),
	)
}

func TestTime(t *testing.T) {
	Test(t,
		// Since runtime duration is non-deterministic, we only have some sanity
		// checks here.
		That("time { echo foo } | var a _ = (all)", "put $a").Puts("foo"),
		That("var duration = ''",
			"time &on-end={|x| set duration = $x } { echo foo } | var out = (all)",
			"put $out", "kind-of $duration").Puts("foo", "number"),
		That("time { fail body } | nop (all)").Throws(FailError{"body"}),
		That("time &on-end={|_| fail on-end } { }").Throws(
			FailError{"on-end"}),

		That("time &on-end={|_| fail on-end } { fail body }").Throws(
			FailError{"body"}),

		thatOutputErrorIsBubbled("time { }"),
	)
}

func TestBenchmark(t *testing.T) {
	var ticks []int64
	testutil.Set(t, TimeNow, func() time.Time {
		if len(ticks) == 0 {
			panic("mock TimeNow called more than len(ticks)")
		}
		v := ticks[0]
		ticks = ticks[1:]
		return time.Unix(v, 0)
	})
	setupTicks := func(ts ...int64) func(*Evaler) {
		return func(_ *Evaler) { ticks = ts }
	}

	Test(t,
		// Default output
		That("benchmark &min-runs=2 &min-time=2s { }").
			WithSetup(setupTicks(0, 1, 1, 3)).
			Prints("1.5s ± 500ms (min 1s, max 2s, 2 runs)\n"),
		// &on-end callback
		That(
			"var f = {|m| put $m[avg] $m[stddev] $m[min] $m[max] $m[runs]}",
			"benchmark &min-runs=2 &min-time=2s &on-end=$f { }").
			WithSetup(setupTicks(0, 1, 1, 3)).
			Puts(1.5, 0.5, 1.0, 2.0, 2),

		// &min-runs determining number of runs
		That("benchmark &min-runs=4 &min-time=0s &on-end={|m| put $m[runs]} { }").
			WithSetup(setupTicks(0, 1, 1, 3, 3, 4, 4, 6)).
			Puts(4),
		// &min-time determining number of runs
		That("benchmark &min-runs=0 &min-time=10s &on-end={|m| put $m[runs]} { }").
			WithSetup(setupTicks(0, 1, 1, 6, 6, 11)).
			Puts(3),

		// &on-run-end
		That("benchmark &min-runs=3 &on-run-end=$put~ &on-end={|m| } { }").
			WithSetup(setupTicks(0, 1, 1, 3, 3, 4)).
			Puts(1.0, 2.0, 1.0),

		// $callable throws exception
		That(
			"var i = 0",
			"benchmark { set i = (+ $i 1); if (== $i 3) { fail failure } }").
			WithSetup(setupTicks(0, 1, 1, 3, 3)).
			Throws(FailError{"failure"}).
			Prints("1.5s ± 500ms (min 1s, max 2s, 2 runs)\n"),
		// $callable throws exception on first run
		That("benchmark { fail failure }").
			WithSetup(setupTicks(0)).
			Throws(FailError{"failure"}).
			Prints( /* nothing */ ""),
		That("benchmark &on-end=$put~ { fail failure }").
			WithSetup(setupTicks(0)).
			Throws(FailError{"failure"}).
			Puts( /* nothing */ ),

		// &on-run-end throws exception
		That("benchmark &on-run-end={|_| fail failure } { }").
			WithSetup(setupTicks(0, 1)).
			Throws(FailError{"failure"}).
			Prints("1s ± 0s (min 1s, max 1s, 1 runs)\n"),

		// &on-run throws exception
		That("benchmark &min-runs=2 &min-time=0s &on-end={|_| fail failure } { }").
			WithSetup(setupTicks(0, 1, 1, 3)).
			Throws(FailError{"failure"}),

		// Option errors
		That("benchmark &min-runs=-1 { }").
			Throws(errs.BadValue{What: "min-runs option",
				Valid: "non-negative integer", Actual: "-1"}),
		That("benchmark &min-time=abc { }").
			Throws(errs.BadValue{What: "min-time option",
				Valid: "duration string", Actual: "abc"}),
		That("benchmark &min-time=-1s { }").
			Throws(errs.BadValue{What: "min-time option",
				Valid: "non-negative duration", Actual: "-1s"}),

		// Test that output error is bubbled. We can't use
		// testOutputErrorIsBubbled here, since the mock TimeNow requires setup.
		That("benchmark &min-runs=0 &min-time=0s { } >&-").
			WithSetup(setupTicks(0, 1)).
			Throws(os.ErrInvalid),
		That("benchmark &min-runs=0 &min-time=0s &on-end=$put~ { } >&-").
			WithSetup(setupTicks(0, 1)).
			Throws(ErrPortDoesNotSupportValueOutput),
	)
}

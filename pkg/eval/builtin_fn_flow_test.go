package eval_test

import (
	"runtime"
	"testing"

	. "src.elv.sh/pkg/eval"

	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestRunParallel(t *testing.T) {
	Test(t,
		That(`run-parallel { put lorem } { echo ipsum }`).
			Puts("lorem").Prints("ipsum\n"),
		That(`run-parallel { } { fail foo }`).Throws(FailError{"foo"}),
	)
}

func TestEach(t *testing.T) {
	Test(t,
		That(`put 1 233 | each $put~`).Puts("1", "233"),
		That(`echo "1\n233" | each $put~`).Puts("1", "233"),
		That(`echo "1\r\n233" | each $put~`).Puts("1", "233"),
		That(`each $put~ [1 233]`).Puts("1", "233"),
		That(`range 10 | each {|x| if (== $x 4) { break }; put $x }`).
			Puts(0, 1, 2, 3),
		That(`range 10 | each {|x| if (== $x 4) { continue }; put $x }`).
			Puts(0, 1, 2, 3, 5, 6, 7, 8, 9),
		That(`range 10 | each {|x| if (== $x 4) { fail haha }; put $x }`).
			Puts(0, 1, 2, 3).Throws(FailError{"haha"}),
		// TODO(xiaq): Test that "each" does not close the stdin.
	)
}

func TestPeach(t *testing.T) {
	// Testing the `peach` builtin is a challenge since, by definition, the order of execution is
	// undefined. Start by validating its options have reasonable values.
	Test(t,
		// Invalid options are handled.
		That(`peach &num-workers=0 {|x| * 2 $x }`).Throws(errs.BadValue{
			What:   "peach &num-workers",
			Valid:  "-1 or > 0",
			Actual: "0",
		}),
		That(`peach &num-workers=-2 {|x| * 2 $x }`).Throws(errs.BadValue{
			What:   "peach &num-workers",
			Valid:  "-1 or > 0",
			Actual: "-2",
		}),

		// Verify the output has the expected values when sorted.
		That(`range 5 | peach {|x| * 2 $x } | order`).Puts(0, 2, 4, 6, 8),

		// Handling of "continue".
		That(`range 5 | peach {|x| if (== $x 2) { continue }; * 2 $x } | order`).
			Puts(0, 2, 6, 8),

		// Test that the order of output does not necessarily match the order of
		// input.
		//
		// Most of the time this effect can be observed without the need of any
		// jitter, but if the system only has one CPU core to execute goroutines
		// (which can happen even when GOMAXPROCS > 1), the scheduling of
		// goroutines can become deterministic. The random jitter fixes that by
		// forcing goroutines to yield the thread and allow other goroutines to
		// execute.
		That(`
			var @in = (range 100)
			while $true {
				var @out = (all $in | peach {|x| sleep (* (rand) 0.01); put $x })
				if (not-eq $in $out) {
					put $true
					break
				}
			}
		`).Puts(true),
		// Verify that exceptions are propagated.
		That(`peach {|x| fail $x } [a]`).
			Throws(FailError{"a"}, "fail $x ", "peach {|x| fail $x } [a]"),
		// Verify that `break` works by terminating the `peach` before the entire sequence is
		// consumed.
		That(`
			var tot = 0
			range 1 101 |
				peach {|x| if (== 50 $x) { break } else { put $x } } |
				< (+ (all)) (+ (range 1 101))
		`).Puts(true),
	)

	// On non-Windows systems we can expect to have `date '+%s'`. This allows us
	// to verify that limiting the number of workers works as well as the
	// parallelism `peach` is meant to provide. Whole second granularity makes
	// this a reliable enough test even in CI environments where the system may
	// have unexpected scheduling latencies.
	//
	// Despite the above these tests require a hack. It reflects the fact that
	// sometimes, due to unexpected scheduling latencies, the elapsed time will
	// be one second longer than expected. TBD is how to make the tests more
	// robust without such hacks. Not to mention reducing the sleep durations so
	// these tests complete more quickly.
	//
	// TODO: Eliminate the dependency on an external `date` command when
	// https://github.com/elves/elvish/issues/1030 is resolved. Thus allowing
	// these tests to also be run on Windows.
	if runtime.GOOS != "windows" {
		Test(t,
			That(`
				var start = (e:date '+%s')
				range 5 | peach &num-workers=-1 {|_|
					sleep 1
					var e = (- (e:date '+%s') $start)
					if (== $e 2) { set e = (num 1) }
					put $e
				}
			`).Puts(1, 1, 1, 1, 1),
		)
		Test(t,
			That(`
				var start = (e:date '+%s')
				range 5 | peach &num-workers=3 {|x|
					sleep 1
					var e = (- (e:date '+%s') $start)
					if (and (< $x 3) (== $e 2)) { set e = (num 1) }
					if (and (> $x 2) (== $e 3)) { set e = (num 2) }
					put $e
				}
			`).Puts(1, 1, 1, 2, 2),
		)
	}
}

func TestFail(t *testing.T) {
	Test(t,
		That("fail haha").Throws(FailError{"haha"}, "fail haha"),
		That("fn f { fail haha }", "fail ?(f)").Throws(
			FailError{"haha"}, "fail haha ", "f"),
		That("fail []").Throws(
			FailError{vals.EmptyList}, "fail []"),
		That("put ?(fail 1)[reason][type]").Puts("fail"),
		That("put ?(fail 1)[reason][content]").Puts("1"),
	)
}

func TestReturn(t *testing.T) {
	Test(t,
		That("return").Throws(Return),
		// Use of return inside fn is tested in TestFn
	)
}

func TestDefer(t *testing.T) {
	Test(t,
		That("{ defer { put a }; put b }").Puts("b", "a"),
		That("defer { }").
			Throws(ErrorWithMessage("defer must be called from within a closure")),
		That("{ defer { fail foo } }").
			Throws(FailError{"foo"}, "fail foo ", "{ defer { fail foo } }"),
		That("{ defer {|x| } }").Throws(
			errs.ArityMismatch{What: "arguments",
				ValidLow: 1, ValidHigh: 1, Actual: 0},
			"defer {|x| } ", "{ defer {|x| } }"),
	)
}

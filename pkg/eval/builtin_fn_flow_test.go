package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestRunParallel(t *testing.T) {
	Test(t,
		That(`run-parallel { put lorem } { echo ipsum }`).
			Puts("lorem").Prints("ipsum\n"),
	)
}

func TestEach(t *testing.T) {
	Test(t,
		That(`put 1 233 | each $put~`).Puts("1", "233"),
		That(`echo "1\n233" | each $put~`).Puts("1", "233"),
		That(`echo "1\r\n233" | each $put~`).Puts("1", "233"),
		That(`each $put~ [1 233]`).Puts("1", "233"),
		That(`range 10 | each [x]{ if (== $x 4) { break }; put $x }`).
			Puts(0, 1, 2, 3),
		That(`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`).
			Puts(0, 1, 2, 3).Throws(AnyError),
		// TODO(xiaq): Test that "each" does not close the stdin.
	)
}

func TestPeach(t *testing.T) {
	// Testing the `peach` builtin is a challenge since, by definition, the order of execution is
	// undefined.
	Test(t,
		// Verify the output has the expected values when sorted.
		That(`range 5 | peach [x]{ * 2 $x } | order`).Puts(0, 2, 4, 6, 8),
		// Verify that successive runs produce the output in different order. This test can
		// theoretically suffer false positives but the vast majority of the time this will produce
		// the expected output in the first iteration. The probability it will produce the same
		// order of output in 100 iterations is effectively zero.
		That(`
			fn f { range 100 | peach $put~ | put [(all)] }
			var x = (f)
			for _ [(range 100)] {
				var y = (f)
				if (not-eq $x $y) {
					put $true
					break
				}
			}
		`).Puts(true),
		// Verify that exceptions are propagated.
		That(`peach [x]{ fail $x } [a]`).
			Throws(FailError{"a"}, "fail $x ", "peach [x]{ fail $x } [a]"),
		// Verify that `break` works by terminating the `peach` before the entire sequence is
		// consumed.
		That(`
			var tot = 0
			range 1 101 |
				peach [x]{ if (== 50 $x) { break } else { put $x } } |
				< (+ (all)) (+ (range 1 101))
		`).Puts(true),
	)
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

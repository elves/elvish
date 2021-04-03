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
			Puts(int64(0), int64(1), int64(2), int64(3)),
		That(`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`).
			Puts(int64(0), int64(1), int64(2), int64(3)).Throws(AnyError),
		// TODO(xiaq): Test that "each" does not close the stdin.
	)
}

// TODO: test peach

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

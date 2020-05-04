package eval

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/eval/errs"
)

func TestBuiltinFnFlow(t *testing.T) {
	Test(t,
		That(`run-parallel { put lorem } { echo ipsum }`).
			Puts("lorem").Prints("ipsum\n"),

		That(`put 1 233 | each $put~`).Puts("1", "233"),
		That(`echo "1\n233" | each $put~`).Puts("1", "233"),
		That(`echo "1\r\n233" | each $put~`).Puts("1", "233"),
		That(`each $put~ [1 233]`).Puts("1", "233"),
		That(`range 10 | each [x]{ if (== $x 4) { break }; put $x }`).
			Puts(0.0, 1.0, 2.0, 3.0),
		That(`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`).
			Puts(0.0, 1.0, 2.0, 3.0).ThrowsAny(),
		// TODO(xiaq): Test that "each" does not close the stdin.
		// TODO: test peach

		That("fail haha").Throws(errors.New("haha"), "fail haha"),
		That("fn f { fail haha }", "fail ?(f)").Throws(
			errors.New("haha"), "fail haha ", "f"),
		That("fail []").Throws(
			errs.BadValue{What: "argument to fail",
				Valid: "string or exception", Actual: "list"},
			"fail []"),

		That(`return`).ThrowsCause(Return),
	)
}

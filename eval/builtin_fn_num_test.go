package eval

import "testing"

func TestBuiltinFnNum(t *testing.T) {
	runTests(t, []Test{
		That("< 1 2 3").Puts(true),
		That("< 1 3 2").Puts(false),
		That("<= 1 1 2").Puts(true),
		That("<= 1 2 1").Puts(false),
		That("== 1 1 1").Puts(true),
		That("== 1 2 1").Puts(false),
		That("!= 1 2 1").Puts(true),
		That("!= 1 1 2").Puts(false),
		That("> 3 2 1").Puts(true),
		That("> 3 1 2").Puts(false),
		That(">= 3 3 2").Puts(true),
		That(">= 3 2 3").Puts(false),

		// TODO test more edge cases
		{"+ 233100 233", want{out: strs("233333")}},
		{"- 233333 233100", want{out: strs("233")}},
		{"- 233", want{out: strs("-233")}},
		{"* 353 661", want{out: strs("233333")}},
		{"/ 233333 353", want{out: strs("661")}},
		{"/ 1 0", want{out: strs("+Inf")}},
		{"^ 16 2", want{out: strs("256")}},
		{"% 23 7", want{out: strs("2")}},
	})
}

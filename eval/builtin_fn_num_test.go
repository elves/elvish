package eval

import "testing"

func TestBuiltinFnNum(t *testing.T) {
	runTests(t, []Test{
		NewTest("< 1 2 3").WantOut(true),
		NewTest("< 1 3 2").WantOut(false),
		NewTest("<= 1 1 2").WantOut(true),
		NewTest("<= 1 2 1").WantOut(false),
		NewTest("== 1 1 1").WantOut(true),
		NewTest("== 1 2 1").WantOut(false),
		NewTest("!= 1 2 1").WantOut(true),
		NewTest("!= 1 1 2").WantOut(false),
		NewTest("> 3 2 1").WantOut(true),
		NewTest("> 3 1 2").WantOut(false),
		NewTest(">= 3 3 2").WantOut(true),
		NewTest(">= 3 2 3").WantOut(false),

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

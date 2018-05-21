package eval

import "testing"

func TestBuiltinFnNum(t *testing.T) {
	Test(t, []TestCase{
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
		That("+ 233100 233").Puts("233333"),
		That("- 233333 233100").Puts("233"),
		That("- 233").Puts("-233"),
		That("* 353 661").Puts("233333"),
		That("/ 233333 353").Puts("661"),
		That("/ 1 0").Puts("+Inf"),
		That("^ 16 2").Puts("256"),
		That("% 23 7").Puts("2"),
	})
}

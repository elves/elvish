package eval

import (
	"math"
	"testing"
)

func TestBuiltinFnNum(t *testing.T) {
	Test(t,
		That("float64 1").Puts(1.0),
		That("float64 (float64 1)").Puts(1.0),

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
		That("+ 233100 233").Puts(233333.0),
		That("- 233333 233100").Puts(233.0),
		That("- 233").Puts(-233.0),
		That("* 353 661").Puts(233333.0),
		That("/ 233333 353").Puts(661.0),
		That("/ 1 0").Puts(math.Inf(1)),
		That("^ 16 2").Puts(256.0),
		That("^ 2 2 3").Puts(256.0),
		That("% 23 7").Puts("2"),
		That("% 1 0").ThrowsAny(),
	)
}

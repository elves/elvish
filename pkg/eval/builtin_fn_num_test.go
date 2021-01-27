package eval_test

import (
	"math"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestFloat64(t *testing.T) {
	Test(t,
		That("float64 1").Puts(1.0),
		That("float64 (float64 1)").Puts(1.0),
	)
}

func TestNumberComparisonCommands(t *testing.T) {
	Test(t,
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
	)
}

func TestArithmeticCommands(t *testing.T) {
	Test(t,
		// TODO test more edge cases
		That("+ 233100 233").Puts(233333.0),
		That("- 233333 233100").Puts(233.0),
		That("- 233").Puts(-233.0),
		That("* 353 661").Puts(233333.0),
		That("/ 233333 353").Puts(661.0),
		That("/ 1 0").Puts(math.Inf(1)),
		That("% 23 7").Puts("2"),
		That("% 1 0").Throws(AnyError),
	)
}

func TestRandint(t *testing.T) {
	Test(t,
		That("randint 1 2").Puts("1"),
		That("i = (randint 10 100); >= $i 10; < $i 100").Puts(true, true),
		That("randint 2 1").Throws(ErrArgs, "randint 2 1"),
		That("randint").Throws(ErrorWithType(errs.ArityMismatch{}), "randint"),
		That("randint 1").Throws(ErrorWithType(errs.ArityMismatch{}), "randint 1"),
		That("randint 1 2 3").Throws(ErrorWithType(errs.ArityMismatch{}), "randint 1 2 3"),
	)
}

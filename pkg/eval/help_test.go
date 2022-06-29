package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestHelp(t *testing.T) {
	Test(t,
		That("help UNKNOWN:").Throws(ErrorWithMessage("could not find documentation for namespace UNKNOWN:")),
		That("help UNKNOWN").Prints(""),
		That(`help`).Prints(MatchingRegexp{Pattern: `^Function .*builtin:help.*:\n\nUsage:`}),
		That(`help help`).Prints(MatchingRegexp{Pattern: `^Function .*builtin:help.*:\n\nUsage:`}),
		That(`help pid`).Prints(MatchingRegexp{Pattern: `^Variable .*builtin:pid`}),
		That(`help builtin:help`).Prints(MatchingRegexp{Pattern: `^Function .*builtin:help.*:\n\nUsage:`}),
		That(`help math:`).Prints(MatchingRegexp{Pattern: `(?s)^\033\[[\d;]+mmath:abs.*?\$number.*?Computes the absolute value.*\$math:pi.*?Approximate value of.*?π`}),
		That(`help &fn math:`).Prints(NonMatchingRegexp{Pattern: `(?s)\$math:pi.*?Approximate value of.*?π`}),
		That(`help &var math:`).Prints(NonMatchingRegexp{Pattern: `(?s)^\033\[[\d;]+mmath:abs.*?\$number.*?Computes the absolute value`}),
		That(`help &search computes`).Prints(MatchingRegexp{Pattern: `(?s)^\033\[[\d;]+mmath:abs.*?\$number.*?Computes the absolute value.*math:tanh.*?Computes the hyperbolic tangent`}),
		That(`help &var &search π`).Prints(MatchingRegexp{Pattern: `(?s)^\033\[[\d;]+m\$math:pi.*?Approximate value of.*?π`}),
		That(`help &var &search computes`).Prints(""),
		That(`help &fn &search π`).Prints(""),
		That(`help &search NaN`).Prints(NonMatchingRegexp{Pattern: `\n\033\[[\d;]+mrepeat`}),
		That(`help &search nan`).Prints(MatchingRegexp{Pattern: `\n\033\[[\d;]+mrepeat`}),
		// Since there is no function or variable named "nan" this should do an implicit search.
		That(`help nan`).Prints(MatchingRegexp{Pattern: `\n\033\[[\d;]+mrepeat`}),
		// A case-sensitive search shouldn't match the builtin:repeat command but should still
		// match division ("/") and a bunch of math: functions.
		That(`help &search NaN`).Prints(NonMatchingRegexp{Pattern: `\n\033\[[\d;]+mrepeat`}),
		That(`help &search NaN`).Prints(MatchingRegexp{Pattern: `(?s)^\033\[[\d;]+m/.*math:acos.*math:trunc`}),
	)
}

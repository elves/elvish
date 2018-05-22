package eval

import "testing"

func TestBuiltinFnFlow(t *testing.T) {
	Test(t,
		That(`run-parallel { put lorem } { echo ipsum }`).Puts(
			"lorem").Prints("ipsum\n"),

		That(`put 1 233 | each $put~`).Puts("1", "233"),
		That(`echo "1\n233" | each $put~`).Puts("1", "233"),
		That(`each $put~ [1 233]`).Puts("1", "233"),
		That(`range 10 | each [x]{ if (== $x 4) { break }; put $x }`).Puts(
			"0", "1", "2", "3"),
		That(`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`).Puts(
			"0", "1", "2", "3").Errors(),
		// TODO: test peach

		That(`fail haha`).Errors(),
		That(`return`).ErrorsWith(Return),
	)
}

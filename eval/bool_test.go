package eval

import "testing"

func TestBool(t *testing.T) {
	runTests(t, []Test{
		NewTest("kind-of $true").WantOutStrings("bool"),
		NewTest("eq $true $true").WantOutBools(true),
		NewTest("eq $true true").WantOutBools(false),
		NewTest("repr $true").WantBytesOutString("$true\n"),
		NewTest("repr $false").WantBytesOutString("$false\n"),
		NewTest("bool $true").WantOutBools(true),
		NewTest("bool $false").WantOutBools(false),
	})
}

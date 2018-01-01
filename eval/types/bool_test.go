package types_test

import (
	"testing"

	"github.com/elves/elvish/eval"
)

func TestBool(t *testing.T) {
	eval.RunTests(t, []eval.Test{
		eval.NewTest("kind-of $true").WantOutStrings("bool"),
		eval.NewTest("eq $true $true").WantOutBools(true),
		eval.NewTest("eq $true true").WantOutBools(false),
		eval.NewTest("repr $true").WantBytesOutString("$true\n"),
		eval.NewTest("repr $false").WantBytesOutString("$false\n"),
		eval.NewTest("bool $true").WantOutBools(true),
		eval.NewTest("bool $false").WantOutBools(false),
	}, eval.NewEvaler)
}

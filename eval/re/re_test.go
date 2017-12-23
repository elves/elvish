package re

import (
	"testing"

	"github.com/elves/elvish/eval"
)

var tests = []eval.Test{
	eval.NewTest("re:match . xyz").WantOutBools(true),
}

func TestRe(t *testing.T) {
	eval.RunTests(t, tests, func() *eval.Evaler {
		ev := eval.NewEvaler()
		ev.Builtin["re"+eval.NsSuffix] = eval.NewRoVariable(Ns())
		return ev
	})
}

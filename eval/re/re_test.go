package re

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/vector"
)

var tests = []eval.Test{
	eval.NewTest("re:match . xyz").WantOutBools(true),
	eval.NewTest("re:match . ''").WantOutBools(false),
	eval.NewTest("re:match '[a-z]' A").WantOutBools(false),

	eval.NewTest("re:find . ab").WantOut(
		newMatch("a", 0, 1, vector.Empty.Cons(newSubmatch("a", 0, 1))),
		newMatch("b", 1, 2, vector.Empty.Cons(newSubmatch("b", 1, 2))),
	),
	eval.NewTest("re:find '[A-Z]([0-9])' 'A1 B2'").WantOut(
		newMatch("A1", 0, 2,
			vector.Empty.Cons(newSubmatch("A1", 0, 2)).Cons(newSubmatch("1", 1, 2))),
		newMatch("B2", 3, 5,
			vector.Empty.Cons(newSubmatch("B2", 3, 5)).Cons(newSubmatch("2", 4, 5))),
	),

	eval.NewTest("re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'").WantOutStrings("baSH and zSH"),
	eval.NewTest("re:replace &literal '(ba|z)sh' '$sh' 'bash and zsh'").WantOutStrings("$sh and $sh"),
	eval.NewTest("re:replace '(ba|z)sh' [x]{ put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'").WantOutStrings("BaSh and ZsH"),

	eval.NewTest("re:split : /usr/sbin:/usr/bin:/bin").WantOutStrings("/usr/sbin", "/usr/bin", "/bin"),
	eval.NewTest("re:split &max=2 : /usr/sbin:/usr/bin:/bin").WantOutStrings("/usr/sbin", "/usr/bin:/bin"),

	eval.NewTest("re:quote a.txt").WantOutStrings(`a\.txt`),
	eval.NewTest("re:quote '(*)'").WantOutStrings(`\(\*\)`),
}

func TestRe(t *testing.T) {
	eval.RunTests(t, tests, func() *eval.Evaler {
		ev := eval.NewEvaler()
		ev.Builtin.SetNs("re", Ns())
		return ev
	})
}

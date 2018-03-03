package re

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/vector"
)

var tests = []eval.Test{
	eval.That("re:match . xyz").Puts(true),
	eval.That("re:match . ''").Puts(false),
	eval.That("re:match '[a-z]' A").Puts(false),

	eval.That("re:find . ab").Puts(
		newMatch("a", 0, 1, vector.Empty.Cons(newSubmatch("a", 0, 1))),
		newMatch("b", 1, 2, vector.Empty.Cons(newSubmatch("b", 1, 2))),
	),
	eval.That("re:find '[A-Z]([0-9])' 'A1 B2'").Puts(
		newMatch("A1", 0, 2,
			vector.Empty.Cons(newSubmatch("A1", 0, 2)).Cons(newSubmatch("1", 1, 2))),
		newMatch("B2", 3, 5,
			vector.Empty.Cons(newSubmatch("B2", 3, 5)).Cons(newSubmatch("2", 4, 5))),
	),

	eval.That("re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'").Puts("baSH and zSH"),
	eval.That("re:replace &literal '(ba|z)sh' '$sh' 'bash and zsh'").Puts("$sh and $sh"),
	eval.That("re:replace '(ba|z)sh' [x]{ put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'").Puts("BaSh and ZsH"),

	eval.That("re:split : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin", "/bin"),
	eval.That("re:split &max=2 : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin:/bin"),

	eval.That("re:quote a.txt").Puts(`a\.txt`),
	eval.That("re:quote '(*)'").Puts(`\(\*\)`),
}

func TestRe(t *testing.T) {
	eval.RunTests(t, tests, func() *eval.Evaler {
		ev := eval.NewEvaler()
		ev.Builtin.AddNs("re", Ns)
		return ev
	})
}

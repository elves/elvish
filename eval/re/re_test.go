package re

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/vector"
)

func TestRe(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("re", Ns) }
	That := eval.That
	eval.TestWithSetup(t, setup,
		That("re:match . xyz").Puts(true),
		That("re:match . ''").Puts(false),
		That("re:match '[a-z]' A").Puts(false),

		// Invalid pattern in re:match
		That("re:match '(' x").Errors(),

		That("re:find . ab").Puts(
			newMatch("a", 0, 1, vector.Empty.Cons(newSubmatch("a", 0, 1))),
			newMatch("b", 1, 2, vector.Empty.Cons(newSubmatch("b", 1, 2))),
		),
		That("re:find '[A-Z]([0-9])' 'A1 B2'").Puts(
			newMatch("A1", 0, 2,
				vector.Empty.Cons(newSubmatch("A1", 0, 2)).Cons(newSubmatch("1", 1, 2))),
			newMatch("B2", 3, 5,
				vector.Empty.Cons(newSubmatch("B2", 3, 5)).Cons(newSubmatch("2", 4, 5))),
		),

		// Invalid pattern in re:find
		That("re:find '(' x").Errors(),

		// Without any flag, finds ax
		That("put (re:find 'a(x|xy)' axy)[text]").Puts("ax"),
		// With &longest, finds axy
		That("put (re:find &longest 'a(x|xy)' axy)[text]").Puts("axy"),
		// TODO: Test &posix

		That("re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'").Puts("baSH and zSH"),
		That("re:replace &literal '(ba|z)sh' '$sh' 'bash and zsh'").Puts("$sh and $sh"),
		That("re:replace '(ba|z)sh' [x]{ put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'").Puts("BaSh and ZsH"),

		// Invalid pattern in re:replace
		That("re:replace '(' x bash").Errors(),
		// Replacement function outputs more than one value
		That("re:replace x [x]{ put a b } xx").Errors(),
		// Replacement function outputs non-string value
		That("re:replace x [x]{ put [] } xx").Errors(),
		// Replacement is not string or function
		That("re:replace x [] xx").Errors(),
		// Replacement is function when &literal is set
		That("re:replace &literal x [_]{ put y } xx").Errors(),

		That("re:split : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin", "/bin"),
		That("re:split &max=2 : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin:/bin"),
		// Invalid pattern in re:split
		That("re:split '(' x").Errors(),

		That("re:quote a.txt").Puts(`a\.txt`),
		That("re:quote '(*)'").Puts(`\(\*\)`),
	)
}

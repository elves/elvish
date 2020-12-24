package re

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/eval/vals"
)

func TestRe(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.Global = eval.NsBuilder{}.AddNs("re", Ns).Ns()
	}
	TestWithSetup(t, setup,
		That("re:match . xyz").Puts(true),
		That("re:match . ''").Puts(false),
		That("re:match '[a-z]' A").Puts(false),

		// Invalid pattern in re:match
		That("re:match '(' x").Throws(AnyError),

		That("re:find . ab").Puts(
			matchStruct{"a", 0, 1, vals.MakeList(submatchStruct{"a", 0, 1})},
			matchStruct{"b", 1, 2, vals.MakeList(submatchStruct{"b", 1, 2})},
		),
		That("re:find '[A-Z]([0-9])' 'A1 B2'").Puts(
			matchStruct{"A1", 0, 2, vals.MakeList(
				submatchStruct{"A1", 0, 2}, submatchStruct{"1", 1, 2})},
			matchStruct{"B2", 3, 5, vals.MakeList(
				submatchStruct{"B2", 3, 5}, submatchStruct{"2", 4, 5})},
		),

		// Access to fields in the match StructMap
		That("put (re:find . a)[text start end groups]").
			Puts("a", "0", "1", vals.MakeList(submatchStruct{"a", 0, 1})),

		// Invalid pattern in re:find
		That("re:find '(' x").Throws(AnyError),

		// Without any flag, finds ax
		That("put (re:find 'a(x|xy)' axy)[text]").Puts("ax"),
		// With &longest, finds axy
		That("put (re:find &longest 'a(x|xy)' axy)[text]").Puts("axy"),
		// TODO: Test &posix

		That("re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'").Puts("baSH and zSH"),
		That("re:replace &literal '(ba|z)sh' '$sh' 'bash and zsh'").Puts("$sh and $sh"),
		That("re:replace '(ba|z)sh' [x]{ put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'").Puts("BaSh and ZsH"),

		// Invalid pattern in re:replace
		That("re:replace '(' x bash").Throws(AnyError),
		// Replacement function outputs more than one value
		That("re:replace x [x]{ put a b } xx").Throws(AnyError),
		// Replacement function outputs non-string value
		That("re:replace x [x]{ put [] } xx").Throws(AnyError),
		// Replacement is not string or function
		That("re:replace x [] xx").Throws(AnyError),
		// Replacement is function when &literal is set
		That("re:replace &literal x [_]{ put y } xx").Throws(AnyError),

		That("re:split : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin", "/bin"),
		That("re:split &max=2 : /usr/sbin:/usr/bin:/bin").Puts("/usr/sbin", "/usr/bin:/bin"),
		// Invalid pattern in re:split
		That("re:split '(' x").Throws(AnyError),

		That("re:quote a.txt").Puts(`a\.txt`),
		That("re:quote '(*)'").Puts(`\(\*\)`),
	)
}

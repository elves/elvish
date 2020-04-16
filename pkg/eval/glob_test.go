package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestGlob_Simple(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	mustMkdirAll("z", "z2")
	mustCreateEmpty("bar", "foo", "ipsum", "lorem")

	Test(t,
		That("put *").Puts("bar", "foo", "ipsum", "lorem", "z", "z2"),
		That("put z*").Puts("z", "z2"),
		That("put ?").Puts("z"),
		That("put ????m").Puts("ipsum", "lorem"),
	)
}

func TestGlob_Recursive(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	mustMkdirAll("1/2/3")
	mustCreateEmpty("a.go", "1/a.go", "1/2/3/a.go")

	Test(t,
		That("put **").Puts("1/2/3/a.go", "1/2/3", "1/2", "1/a.go", "1", "a.go"),
		That("put **.go").Puts("1/2/3/a.go", "1/a.go", "a.go"),
		That("put 1**.go").Puts("1/2/3/a.go", "1/a.go"),
	)
}

func TestGlob_NoMatch(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	Test(t,
		That("put a/b/nonexistent*").ThrowsCause(ErrWildcardNoMatch),
		That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),
	)
}

func TestGlob_MatchHidden(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	mustMkdirAll("d", ".d")
	mustCreateEmpty("a", ".a", "d/a", "d/.a", ".d/a", ".d/.a")

	Test(t,
		That("put *").Puts("a", "d"),
		That("put *[match-hidden]").Puts(".a", ".d", "a", "d"),
		That("put *[match-hidden]/*").Puts(".d/a", "d/a"),
		That("put */*[match-hidden]").Puts("d/.a", "d/a"),
		That("put *[match-hidden]/*[match-hidden]").Puts(
			".d/.a", ".d/a", "d/.a", "d/a"),
	)
}

func TestGlob_SetAndRange(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	mustCreateEmpty("a1", "a2", "b1", "c1", "ipsum", "lorem")

	Test(t,
		That("put ?[set:ab]*").Puts("a1", "a2", "b1"),
		That("put ?[range:a-c]*").Puts("a1", "a2", "b1", "c1"),
		That("put ?[range:a~c]*").Puts("a1", "a2", "b1"),
		That("put *[range:a-z]").Puts("ipsum", "lorem"),
	)

}

func TestGlob_But(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	mustCreateEmpty("bar", "foo", "ipsum", "lorem")

	Test(t,
		// Nonexistent files can also be excluded
		That("put *[but:foobar][but:ipsum]").Puts("bar", "foo", "lorem"),
	)
}

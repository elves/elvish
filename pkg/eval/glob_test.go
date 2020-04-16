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

func TestGlob_NoMatch(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	Test(t,
		That("put a/b/nonexistent*").ThrowsCause(ErrWildcardNoMatch),
		That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),
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

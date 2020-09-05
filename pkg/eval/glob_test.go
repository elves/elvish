package eval_test

import (
	"math/rand"
	"sort"
	"testing"

	. "github.com/elves/elvish/pkg/eval"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/testutil"
)

func TestGlob_Simple(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	filesToCreate := []string{"a1", "a2", "b1", "c1", "ipsum", "lorem", "z2"}
	dirsToCreate := []string{"a", "d", "z"}

	// Randomly permute the list of file names to help ensure we detect when
	// the glob expansion is not correctly sorted. This is further reinforced
	// by dirsToCreate containing names that should be interleaved in regular
	// files in a `*` glob expansion.
	rand.Shuffle(len(filesToCreate), func(i, j int) {
		filesToCreate[i], filesToCreate[j] = filesToCreate[j], filesToCreate[i]
	})
	rand.Shuffle(len(dirsToCreate), func(i, j int) {
		dirsToCreate[i], dirsToCreate[j] = dirsToCreate[j], dirsToCreate[i]
	})
	testutil.MustCreateEmpty(filesToCreate...)
	testutil.MustMkdirAll(dirsToCreate...)

	all := append(filesToCreate, dirsToCreate...)
	sort.Strings(all)
	wantAll := make([]interface{}, len(all))
	for i, v := range all {
		wantAll[i] = v
	}

	Test(t,
		That("put *").Puts(wantAll...),
		That("put z*").Puts("z", "z2"),
		That("put ?").Puts("a", "d", "z"),
		That("put ????m").Puts("ipsum", "lorem"),
	)
}

func TestGlob_Recursive(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustMkdirAll("1/2/3")
	testutil.MustCreateEmpty("a.go", "1/a.go", "1/2/3/a.go")

	Test(t,
		That("put **").Puts("1", "1/2", "1/2/3", "1/2/3/a.go", "1/a.go", "a.go"),
		That("put **.go").Puts("1/2/3/a.go", "1/a.go", "a.go"),
		That("put 1**.go").Puts("1/2/3/a.go", "1/a.go"),
	)
}

func TestGlob_NoMatch(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	Test(t,
		That("put a/b/nonexistent*").Throws(ErrWildcardNoMatch),
		That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),
	)
}

func TestGlob_MatchHidden(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustMkdirAll("d", ".d")
	testutil.MustCreateEmpty("a", ".a", "d/a", "d/.a", ".d/a", ".d/.a")

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
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustCreateEmpty("a1", "a2", "b1", "c1", "ipsum", "lorem")

	Test(t,
		That("put ?[set:ab]*").Puts("a1", "a2", "b1"),
		That("put ?[range:a-c]*").Puts("a1", "a2", "b1", "c1"),
		That("put ?[range:a~c]*").Puts("a1", "a2", "b1"),
		That("put *[range:a-z]").Puts("ipsum", "lorem"),
	)

}

func TestGlob_But(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustCreateEmpty("bar", "foo", "ipsum", "lorem")

	Test(t,
		// Nonexistent files can also be excluded
		That("put *[but:foobar][but:ipsum]").Puts("bar", "foo", "lorem"),
	)
}

func TestGlob_Type(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustMkdirAll("d1", "d2", ".d", "b/c")
	testutil.MustCreateEmpty("bar", "foo", "ipsum", "lorem", "d1/f1", "d2/fm")

	Test(t,
		That("put **[type:dir]").Puts("b", "b/c", "d1", "d2"),
		That("put **[type:regular]").Puts("bar", "d1/f1", "d2/fm", "foo", "ipsum", "lorem"),
		That("put **[type:regular]m").Puts("d2/fm", "ipsum", "lorem"),
		That("put **[type:dir]f*[type:regular]").Throws(ErrMultipleTypeModifiers),
		That("put **[type:unknown]").Throws(ErrUnknownTypeModifier),
	)
}

// This test is unusual. It exists solely to verify that duplicate path names
// that can result from globs such as `?[set:aeoiu digit]` both sort the path
// names and remove duplicates. See the code in `func (op compoundOp) exec(fm
// *Frame) ([]interface{}, error)` in pkg/eval/compile_value.go that produces
// unique, sorted, path names from such a glob expansion.
func TestGlob_MultipleMatchers(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	testutil.MustCreateEmpty("x", "a", "bar", "1")

	Test(t,
		That("put ?[set:aeoiu digit]").Puts("1", "a"),
	)
}

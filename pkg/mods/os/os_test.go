package os

import (
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

var removePaths = testutil.Dir{
	"a": "",
	"d": testutil.Dir{
		"b": "",
		"e": testutil.Dir{
			"f": "",
		},
		"c": "",
	},
}

var removeSymlinks = []struct {
	path   string
	target string
}{
	{"s-bad", "/argle/bargle"},
	{"d/s-f", "b"},
}

func TestPathRemove(t *testing.T) {
	tmpdir := testutil.InTempDir(t)
	testutil.ApplyDir(removePaths)
	for _, link := range removeSymlinks {
		err := os.Symlink(link.target, link.path)
		if err != nil {
			// Creating symlinks requires a special permission on Windows. If
			// the user doesn't have that permission, just skip the whole test.
			t.Skip(err)
		}
	}

	// Note that the order of these tests is important as subsequent tests rely
	// on prior tests having removed specific files.
	TestWithSetup(t, importModules,
		That("os:remove").DoesNothing(),
		That("os:remove does-not-exist").Throws(ErrorWithType(&os.PathError{})),
		That("os:remove &missing-ok does-not-exist").DoesNothing(),
		That("os:remove "+filepath.Join(tmpdir, "d", "e")).Throws(ErrorWithType(&os.PathError{})),
		That("put d/**").Puts("d/e/f", "d/b", "d/c", "d/e", "d/s-f"),
		That("os:remove &recursive "+filepath.Join(tmpdir, "d", "e")).DoesNothing(),
		That("put d/**").Puts("d/b", "d/c", "d/s-f"),
		That("os:remove &recursive *").DoesNothing(),
		That("put **").Throws(eval.ErrWildcardNoMatch),
	)
}

func importModules(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("os", Ns))
}

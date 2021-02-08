package path

import (
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

var testDir = testutil.Dir{
	"d1": testutil.Dir{
		"f": testutil.Symlink{filepath.Join("d2", "f")},
		"d2": testutil.Dir{
			"empty": "",
			"f":     "",
			"g":     testutil.Symlink{"f"},
		},
	},
	"s1": testutil.Symlink{filepath.Join("d1", "d2")},
}

func TestPath(t *testing.T) {
	tmpdir, cleanup := testutil.InTestDir()
	defer cleanup()
	t.Log("test dir", tmpdir)
	testutil.ApplyDir(testDir)

	absPath, err := filepath.Abs("a/b/c.png")
	if err != nil {
		panic("unable to convert a/b/c.png to an absolute path")
	}

	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("path", Ns).Ns())
	}
	TestWithSetup(t, setup,
		// This block of tests is not meant to be comprehensive. Their primary purpose is to simply
		// ensure the Elvish command is correctly mapped to the relevant Go function. We assume the
		// Go function behaves correctly.
		That(`path:abs a/b/c.png`).Puts(absPath),
		That(`path:base a/b/d.png`).Puts("d.png"),
		That(`path:clean ././x`).Puts("x"),
		That(`path:clean a/b/.././c`).Puts(filepath.Join("a", "c")),
		That(`path:dir a/b/d.png`).Puts(filepath.Join("a", "b")),
		That(`path:ext a/b/e.png`).Puts(".png"),
		That(`path:ext a/b/s`).Puts(""),
		That(`path:is-abs a/b/s`).Puts(false),
		That(`path:is-abs `+absPath).Puts(true),
		That(`path:eval-symlinks d1/d2`).Puts(filepath.Join("d1", "d2")),
		That(`path:eval-symlinks d1/d2/f`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1`).Puts(filepath.Join("d1", "d2")),
		That(`path:eval-symlinks d1/f`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1/g`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1/empty`).Puts(filepath.Join("d1", "d2", "empty")),

		// Elvish `path:` module functions that are not trivial wrappers around a Go stdlib function
		// should have comprehensive tests below this comment.
		That(`path:is-dir a/b/s`).Puts(false),
		That(`path:is-dir `+tmpdir).Puts(true),
		That(`path:is-dir s1`).Puts(false),
		That(`path:is-regular a/b/s`).Puts(false),
		That(`path:is-regular `+tmpdir).Puts(false),
		That(`path:is-regular d1/f`).Puts(false),
		That(`path:is-regular d1/d2/f`).Puts(true),
		That(`path:is-regular s1/f`).Puts(true),
	)
}

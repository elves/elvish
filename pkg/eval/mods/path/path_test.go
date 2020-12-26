package path

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/eval"
	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/testutil"
)

func TestPath(t *testing.T) {
	tmpdir, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.Must(os.MkdirAll("d1/d2", 0700))
	testutil.Must(os.Symlink("d1", "s1"))
	absPath, err := filepath.Abs("a/b/c.png")
	if err != nil {
		panic("unable to convert a/b/c.png to an absolute path")
	}

	setup := func(ev *eval.Evaler) {
		ev.Global = eval.NsBuilder{}.AddNs("path", Ns).Ns()
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
		That(`path:real s1/d2`).Puts(filepath.Join("d1", "d2")),

		// Elvish `path:` module functions that are not trivial wrappers around a Go stdlib function
		// should have comprehensive tests below this comment.
		That(`path:is-dir a/b/s`).Puts(false),
		That(`path:is-dir `+tmpdir).Puts(true),
		That(`path:is-dir s1`).Puts(true),
	)
}

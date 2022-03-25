package edit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

// High-level sanity test.

func TestHighlighter(t *testing.T) {
	f := setup(t)

	feedInput(f.TTYCtrl, "put $true")
	f.TestTTY(t,
		"~> put $true", Styles,
		"   vvv $$$$$", term.DotHere,
	)

	feedInput(f.TTYCtrl, "x")
	f.TestTTY(t,
		"~> put $truex", Styles,
		"   vvv ??????", term.DotHere, "\n",
		"compilation error: 4-10 in [interactive]: variable $truex not found",
	)
}

// Fine-grained tests against the highlighter.

func TestCheck(t *testing.T) {
	ev := eval.NewEvaler()
	ev.ExtendGlobal(eval.BuildNs().AddVar("good", vars.FromInit(0)))

	tt.Test(t, tt.Fn("check", check), tt.Table{
		Args(ev, mustParse("")).Rets(noError),
		Args(ev, mustParse("echo $good")).Rets(noError),
		// TODO: Check the range of the returned error
		Args(ev, mustParse("echo $bad")).Rets(anyError),
	})
}

type anyErrorMatcher struct{}

func (anyErrorMatcher) Match(ret tt.RetValue) bool {
	err, _ := ret.(error)
	return err != nil
}

var (
	noError  = error(nil)
	anyError anyErrorMatcher
)

const colonInFilenameOk = runtime.GOOS != "windows"

func TestMakeHasCommand(t *testing.T) {
	ev := eval.NewEvaler()

	// Set up global functions and modules in the evaler.
	goodFn := eval.NewGoFn("good", func() {})
	ev.ExtendGlobal(eval.BuildNs().
		AddFn("good", goodFn).
		AddNs("a",
			eval.BuildNs().
				AddFn("good", goodFn).
				AddNs("b", eval.BuildNs().AddFn("good", goodFn))))

	// Set up environment.
	testDir := testutil.InTempDir(t)
	testutil.Setenv(t, env.PATH, filepath.Join(testDir, "bin"))
	if runtime.GOOS == "windows" {
		testutil.Unsetenv(t, env.PATHEXT) // force default value
	}

	// Set up a directory in PATH.
	mustMkdirAll("bin")
	mustMkExecutable("bin/external")
	mustMkExecutable("bin/@external")
	if colonInFilenameOk {
		mustMkExecutable("bin/ex:tern:al")
	}

	// Set up a directory not in PATH.
	mustMkdirAll("a/b/c")
	mustMkExecutable("a/b/c/executable")

	tt.Test(t, tt.Fn("hasCommand", hasCommand), tt.Table{
		// Builtin special form
		Args(ev, "if").Rets(true),

		// Builtin function
		Args(ev, "put").Rets(true),

		// User-defined function
		Args(ev, "good").Rets(true),

		// Function in modules
		Args(ev, "a:good").Rets(true),
		Args(ev, "a:b:good").Rets(true),
		Args(ev, "a:bad").Rets(false),
		Args(ev, "a:b:bad").Rets(false),

		// Non-searching directory and external
		Args(ev, "./a").Rets(true),
		Args(ev, "a/b").Rets(true),
		Args(ev, "a/b/c/executable").Rets(true),
		Args(ev, "./bad").Rets(false),
		Args(ev, "a/bad").Rets(false),

		// External in PATH
		Args(ev, "external").Rets(true),
		Args(ev, "@external").Rets(true),
		Args(ev, "ex:tern:al").Rets(colonInFilenameOk),
		// With explicit e:
		Args(ev, "e:external").Rets(true),
		Args(ev, "e:bad-external").Rets(false),

		// Non-existent
		Args(ev, "bad").Rets(false),
		Args(ev, "a:").Rets(false),
	})
}

func mustParse(src string) parse.Tree {
	tree, err := parse.Parse(parse.SourceForTest(src), parse.Config{})
	if err != nil {
		panic(err)
	}
	return tree
}

func mustMkdirAll(path string) {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(err)
	}
}

func mustMkExecutable(path string) {
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	err := os.WriteFile(path, nil, 0700)
	if err != nil {
		panic(err)
	}
}

package cliedit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
	"github.com/elves/elvish/util"
)

// High-level sanity test.

func TestHighlighter(t *testing.T) {
	_, cleanupDir := eval.InTempHome()
	defer cleanupDir()
	_, ttyCtrl, _, cleanup := setupStarted()
	defer cleanup()

	feedInput(ttyCtrl, "put $true")
	wantBuf1 := bb().
		WriteStyled(styled.MarkLines(
			"~> put $true", styles,
			"   ggg vvvvv",
		)).
		SetDotToCursor().
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf1)

	feedInput(ttyCtrl, "x")
	wantBuf2 := bb().
		WriteStyled(styled.MarkLines(
			"~> put $truex", styles,
			"   ggg eeeeee",
		)).
		SetDotToCursor().
		Newline().
		WritePlain("compilation error: 4-10 in [tty]: variable $truex not found").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf2)
}

// Fine-grained tests against the highlighter.

func TestCheck(t *testing.T) {
	ev := eval.NewEvaler()
	ev.Global.Add("good", vars.FromInit(0))

	tt.Test(t, tt.Fn("check", check), tt.Table{
		tt.Args(ev, mustParse("")).Rets(noError),
		tt.Args(ev, mustParse("echo $good")).Rets(noError),
		// TODO: Check the range of the returned error
		tt.Args(ev, mustParse("echo $bad")).Rets(anyError),
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
	ev.Global.AddFn("good", goodFn)
	aNs := eval.Ns{}.AddFn("good", goodFn)
	bNs := eval.Ns{}.AddFn("good", goodFn)
	aNs.AddNs("b", bNs)
	ev.Global.AddNs("a", aNs)

	// Set up environment.
	testDir, cleanup := util.InTestDir()
	defer cleanup()
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	if runtime.GOOS == "windows" {
		oldPathExt := os.Getenv("PATHEXT")
		defer os.Setenv("PATHEXT", oldPathExt)
		// Forces default value
		os.Setenv("PATHEXT", "")
	}

	// Set up a directory in PATH.
	os.Setenv("PATH", filepath.Join(testDir, "bin"))
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
		tt.Args(ev, "if").Rets(true),
		// Builtin function
		tt.Args(ev, "put").Rets(true),
		// User-defined function
		tt.Args(ev, "good").Rets(true),
		// Function in modules
		tt.Args(ev, "a:good").Rets(true),
		tt.Args(ev, "a:b:good").Rets(true),

		// Non-searching directory and external
		tt.Args(ev, "./a").Rets(true),
		tt.Args(ev, "a/b").Rets(true),
		tt.Args(ev, "a/b/c/executable").Rets(true),

		// External in PATH
		tt.Args(ev, "external").Rets(true),
		tt.Args(ev, "@external").Rets(true),
		tt.Args(ev, "ex:tern:al").Rets(colonInFilenameOk),

		// Non-existent
		tt.Args(ev, "bad").Rets(false),
		tt.Args(ev, "a:").Rets(false),
		tt.Args(ev, "a:bad").Rets(false),
		tt.Args(ev, "a:b:bad").Rets(false),
		tt.Args(ev, "./bad").Rets(false),
		tt.Args(ev, "a/bad").Rets(false),
	})
}

func mustParse(src string) *parse.Chunk {
	n, err := parse.AsChunk("[test]", src)
	if err != nil {
		panic(err)
	}
	return n
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
	err := ioutil.WriteFile(path, nil, 0700)
	if err != nil {
		panic(err)
	}
}

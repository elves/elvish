package newedit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/tt"
	"github.com/elves/elvish/util"
)

type anyErrorMatcher struct{}

func (anyErrorMatcher) Match(ret tt.RetValue) bool {
	err, _ := ret.(error)
	return err != nil
}

var (
	noError  error
	anyError anyErrorMatcher
)

func TestMakeCheck(t *testing.T) {
	ev := eval.NewEvaler()
	ev.Global.Add("good", vars.FromInit(0))
	check := makeCheck(ev)

	tt.Test(t, tt.Fn("check", check), tt.Table{
		tt.Args(mustParse("")).Rets(noError),
		tt.Args(mustParse("echo $good")).Rets(noError),
		// TODO: Check the range of the returned error
		tt.Args(mustParse("echo $bad")).Rets(anyError),
	})
}

const colonInFilenameOk = runtime.GOOS != "windows"

func TestMakeHasCommand(t *testing.T) {
	ev := eval.NewEvaler()
	hasCommand := makeHasCommand(ev)

	// Set up global functions and modules in the evaler.
	goodFn := eval.NewBuiltinFn("good", func() {})
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
		tt.Args("if").Rets(true),
		// Builtin function
		tt.Args("put").Rets(true),
		// User-defined function
		tt.Args("good").Rets(true),
		// Function in modules
		tt.Args("a:good").Rets(true),
		tt.Args("a:b:good").Rets(true),

		// Non-searching directory and external
		tt.Args("./a").Rets(true),
		tt.Args("a/b").Rets(true),
		tt.Args("a/b/c/executable").Rets(true),

		// External in PATH
		tt.Args("external").Rets(true),
		tt.Args("@external").Rets(true),
		tt.Args("ex:tern:al").Rets(colonInFilenameOk),

		// Non-existent
		tt.Args("bad").Rets(false),
		tt.Args("a:bad").Rets(false),
		tt.Args("a:b:bad").Rets(false),
		tt.Args("./bad").Rets(false),
		tt.Args("a/bad").Rets(false),
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
		path = path + ".exe"
	}
	err := ioutil.WriteFile(path, nil, 0700)
	if err != nil {
		panic(err)
	}
}

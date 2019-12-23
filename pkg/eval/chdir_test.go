package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
)

type testAddDirer func(string, float64) error

func (t testAddDirer) AddDir(dir string, weight float64) error {
	return t(dir, weight)
}

func TestChdir(t *testing.T) {
	dst, cleanup := util.TestDir()
	defer cleanup()

	ev := NewEvaler()

	argDirInBefore, argDirInAfter := "", ""
	ev.AddBeforeChdir(func(dir string) { argDirInBefore = dir })
	ev.AddAfterChdir(func(dir string) { argDirInAfter = dir })

	back := saveWd()
	defer back()

	err := ev.Chdir(dst)

	if err != nil {
		t.Errorf("Chdir => error %v", err)
	}
	if envPwd := os.Getenv("PWD"); envPwd != dst {
		t.Errorf("$PWD is %q after Chdir, want %q", envPwd, dst)
	}

	if argDirInBefore != dst {
		t.Errorf("Chdir called before-hook with %q, want %q",
			argDirInBefore, dst)
	}
	if argDirInAfter != dst {
		t.Errorf("Chdir called before-hook with %q, want %q",
			argDirInAfter, dst)
	}
}

func TestChdirElvishHooks(t *testing.T) {
	dst, cleanup := util.TestDir()
	defer cleanup()

	back := saveWd()
	defer back()

	Test(t,
		That(`
			dir-in-before dir-in-after = '' ''
			@before-chdir = [dst]{ dir-in-before = $dst }
			@after-chdir  = [dst]{ dir-in-after  = $dst }
			cd `+parse.Quote(dst)+`
			put $dir-in-before $dir-in-after
			`).Puts(dst, dst),
	)
}

func TestChdirError(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	ev := NewEvaler()
	err := ev.Chdir("i/dont/exist")
	if err == nil {
		t.Errorf("Chdir => no error when dir does not exist")
	}
}

// Saves the current working directory, and returns a function for returning to
// it.
func saveWd() func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return func() {
		err := os.Chdir(wd)
		if err != nil {
			panic(err)
		}
	}
}

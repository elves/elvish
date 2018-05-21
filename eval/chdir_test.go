package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type testAddDirer func(string, float64) error

func (t testAddDirer) AddDir(dir string, weight float64) error {
	return t(dir, weight)
}

func TestChdir(t *testing.T) {
	inWithTempDir(func(pwd, dst string) {
		ev := NewEvaler()

		argDirInBefore, argDirInAfter := "", ""
		ev.AddBeforeChdir(func(dir string) { argDirInBefore = dir })
		ev.AddAfterChdir(func(dir string) { argDirInAfter = dir })

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
	})
}

func TestChdirElvishHooks(t *testing.T) {
	inWithTempDir(func(pwd, dst string) {
		Test(t, []TestCase{
			That(`
			dir-in-before dir-in-after = '' ''
			@before-chdir = [dst]{ dir-in-before = $dst }
			@after-chdir  = [dst]{ dir-in-after  = $dst }
			cd `+parse.Quote(dst)+`
			put $dir-in-before $dir-in-after
			`).Puts(dst, dst),
		})
	})
}

func TestChdirError(t *testing.T) {
	util.InTempDir(func(pwd string) {
		ev := NewEvaler()
		err := ev.Chdir("i/dont/exist")
		if err == nil {
			t.Errorf("Chdir => no error when dir does not exist")
		}
	})
}

func inWithTempDir(f func(pwd, other string)) {
	util.InTempDir(func(pwd string) {
		util.WithTempDir(func(other string) {
			f(pwd, other)
		})
	})
}

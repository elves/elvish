package eval_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/env"
	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestChdir(t *testing.T) {
	dst := testutil.TempDir(t)

	ev := NewEvaler()

	argDirInBefore, argDirInAfter := "", ""
	ev.BeforeChdir = append(ev.BeforeChdir, func(dir string) { argDirInBefore = dir })
	ev.AfterChdir = append(ev.AfterChdir, func(dir string) { argDirInAfter = dir })

	back := saveWd()
	defer back()

	err := ev.Chdir(dst)

	if err != nil {
		t.Errorf("Chdir => error %v", err)
	}
	if envPwd := os.Getenv(env.PWD); envPwd != dst {
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

func TestChdirError(t *testing.T) {
	testutil.InTempDir(t)

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
		must.Chdir(wd)
	}
}

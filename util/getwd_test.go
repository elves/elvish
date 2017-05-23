package util

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetwd(t *testing.T) {
	InTempDir(func(tmpdir string) {
		// On some systems /tmp is a symlink.
		tmpdir, err := filepath.EvalSymlinks(tmpdir)
		if err != nil {
			panic(err)
		}
		if gotwd := Getwd(); gotwd != tmpdir {
			t.Errorf("Getwd() -> %v, want %v", gotwd, tmpdir)
		}

		// Override $HOME to trick GetHome.
		os.Setenv("HOME", tmpdir)

		if gotwd := Getwd(); gotwd != "~" {
			t.Errorf("Getwd() -> %v, want ~", gotwd)
		}

		mustOK(os.Mkdir("a", 0700))
		mustOK(os.Chdir("a"))
		if gotwd := Getwd(); gotwd != "~/a" {
			t.Errorf("Getwd() -> %v, want ~/a", gotwd)
		}

		// On macOS os.Getwd will still return the old path name in face of
		// directory being removed. Hence we only test this on Linux.
		// TODO(xiaq): Check the behavior on other BSDs and relax this condition
		// if possible.
		if runtime.GOOS == "linux" {
			if gotwd := Getwd(); gotwd != "?" {
				t.Errorf("Getwd() -> %v, want ?", gotwd)
			}
		}
	})
}

func mustOK(err error) {
	if err != nil {
		panic(err)
	}
}

package util

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetwd(t *testing.T) {
	tmpdir, cleanup := InTestDir()
	defer cleanup()

	// On some systems /tmp is a symlink.
	tmpdir, err := filepath.EvalSymlinks(tmpdir)
	if err != nil {
		panic(err)
	}
	// Override $HOME to make sure that tmpdir is not abbreviatable.
	os.Setenv("HOME", "/does/not/exist")
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
	if gotwd := Getwd(); gotwd != filepath.Join("~", "a") {
		t.Errorf("Getwd() -> %v, want ~/a", gotwd)
	}

	// On macOS os.Getwd will still return the old path name in face of
	// directory being removed. Hence we only test this on Linux.
	// TODO(xiaq): Check the behavior on other BSDs and relax this condition
	// if possible.
	if runtime.GOOS == "linux" {
		mustOK(os.Remove(path.Join(tmpdir, "a")))
		if gotwd := Getwd(); gotwd != "?" {
			t.Errorf("Getwd() -> %v, want ?", gotwd)
		}
	}
}

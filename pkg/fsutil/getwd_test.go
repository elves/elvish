package fsutil

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/env"
	"github.com/elves/elvish/pkg/testutil"
)

func TestGetwd(t *testing.T) {
	tmpdir, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.Must(os.Mkdir("a", 0700))

	// On some systems /tmp is a symlink.
	tmpdir, err := filepath.EvalSymlinks(tmpdir)
	if err != nil {
		panic(err)
	}

	var tests = []struct {
		name   string
		home   string
		chdir  string
		wantWd string
	}{
		{"wd outside HOME not abbreviated", "/does/not/exist", tmpdir, tmpdir},

		{"wd at HOME abbreviated", tmpdir, tmpdir, "~"},
		{"wd inside HOME abbreviated", tmpdir, tmpdir + "/a", filepath.Join("~", "a")},

		{"wd not abbreviated when HOME is slash", "/", tmpdir, tmpdir},
	}

	oldHome := os.Getenv(env.HOME)
	defer os.Setenv(env.HOME, oldHome)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(env.HOME, test.home)
			testutil.MustChdir(test.chdir)
			if gotWd := Getwd(); gotWd != test.wantWd {
				t.Errorf("Getwd() -> %v, want %v", gotWd, test.wantWd)
			}
		})
	}

	// Remove the working directory, and test that Getwd returns "?".
	//
	// This test is now only enabled on Linux, where os.Getwd returns an error
	// when the working directory has been removed. Other operating systems may
	// return the old path even if it is now invalid.
	//
	// TODO(xiaq): Check all the supported operating systems and see which ones
	// have the same behavior as Linux. So far only macOS has been checked.
	if runtime.GOOS == "linux" {
		wd := path.Join(tmpdir, "a")
		testutil.MustChdir(wd)
		testutil.Must(os.Remove(wd))
		if gotwd := Getwd(); gotwd != "?" {
			t.Errorf("Getwd() -> %v, want ?", gotwd)
		}
	}
}

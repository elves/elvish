package fsutil

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestGetwd(t *testing.T) {
	tmpdir := testutil.InTempDir(t)
	must.OK(os.Mkdir("a", 0700))

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

	testutil.SaveEnv(t, env.HOME)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(env.HOME, test.home)
			must.Chdir(test.chdir)
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
		must.Chdir(wd)
		must.OK(os.Remove(wd))
		if gotwd := Getwd(); gotwd != "?" {
			t.Errorf("Getwd() -> %v, want ?", gotwd)
		}
	}
}

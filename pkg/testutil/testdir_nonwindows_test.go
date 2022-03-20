//go:build !windows

package testutil

import (
	"os"
	"testing"
)

func TestApplyDir_CreatesFileWithPerm(t *testing.T) {
	InTempDir(t)

	ApplyDir(Dir{
		// For some unknown reason, termux on Android does not set the
		// group and other permission bits correctly, so we use 700 here.
		"a": File{0700, "a content"},
	})

	testFileContent(t, "a", "a content")
	testFilePerm(t, "a", 0700)
}

func testFilePerm(t *testing.T, filename string, wantPerm os.FileMode) {
	t.Helper()
	info, err := os.Stat(filename)
	if err != nil {
		t.Errorf("Could not stat %v: %v", filename, err)
		return
	}
	if perm := info.Mode().Perm(); perm != wantPerm {
		t.Errorf("File %v has perm %o, want %o", filename, perm, wantPerm)
		wd, err := os.Getwd()
		if err == nil {
			t.Logf("pwd is %v", wd)
		}
	}
}

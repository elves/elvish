package lscolors

import (
	"os"
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestDetermineFeature(t *testing.T) {
	test := func(fname string, mh bool, wantedFeature feature) {
		t.Helper()
		feature, err := determineFeature(fname, mh)
		if err != nil {
			t.Errorf("determineFeature(%q, %v) returns error %v, want no error",
				fname, mh, err)
		}
		if feature != wantedFeature {
			t.Errorf("determineFeature(%q, %v) returns feature %v, want %v",
				fname, mh, feature, wantedFeature)
		}
	}

	_, cleanup := util.InTestDir()
	defer cleanup()

	create("a", 0600)
	// Regular file.
	test("a", true, featureRegular)

	// Symlink.
	err := os.Symlink("a", "symlink")
	if err != nil {
		t.Logf("Failed to create symlink: %v; skipping symlink test", err)
	} else {
		test("symlink", true, featureSymlink)
	}

	// Broken symlink.
	err = os.Symlink("aaaa", "bad-symlink")
	if err != nil {
		t.Logf("Failed to create bad symlink: %v; skipping bad symlink test", err)
	} else {
		test("bad-symlink", true, featureOrphanedSymlink)
	}

	if runtime.GOOS != "windows" {
		// Multiple hard links.
		err := os.Link("a", "a2")
		if err != nil {
			t.Logf("Failed to create hard link: %v; skipping hard link test", err)
		} else {
			test("a", true, featureMultiHardLink)
		}
	}

	// Don't test for multiple hard links.
	test("a", false, featureRegular)

	// Setuid and Setgid.
	// TODO(xiaq): Fails.
	/*
		create("su", os.ModeSetuid)
		test("su", true, featureSetuid)
		create("sg", os.ModeSetgid)
		test("sg", true, featureSetgid)
	*/

	if runtime.GOOS != "windows" {
		// Executable.
		create("xu", 0100)
		create("xg", 0010)
		create("xo", 0001)
		test("xu", true, featureExecutable)
		test("xg", true, featureExecutable)
		test("xo", true, featureExecutable)
	}
}

func create(fname string, perm os.FileMode) {
	f, err := os.OpenFile(fname, os.O_CREATE, perm)
	if err != nil {
		panic(err)
	}
	// The following call works around a termux bug where os.OpenFile does not
	// set the permission bits for group and other correctly.
	f.Chmod(perm)
	f.Close()
}

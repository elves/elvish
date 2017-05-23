package edit

import (
	"os"
	"testing"

	"github.com/elves/elvish/util"
)

func create(fname string, perm os.FileMode) {
	f, err := os.OpenFile(fname, os.O_CREATE, perm)
	if err != nil {
		panic(err)
	}
	f.Close()
}

func TestDetermineFeature(t *testing.T) {
	test := func(fname string, mh bool, wantedFeature fileFeature) {
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

	util.InTempDir(func(string) {

		create("a", 0600)
		// Regular file.
		test("a", true, featureRegular)

		// Symlink.
		os.Symlink("a", "symlink")
		test("symlink", true, featureSymlink)

		// Broken symlink.
		os.Symlink("aaaa", "bad-symlink")
		test("bad-symlink", true, featureOrphanedSymlink)

		// Multiple hard links.
		os.Link("a", "a2")
		test("a", true, featureMultiHardLink)

		// Don't test for multiple hard links.
		test("a", false, featureRegular)

		// Setuid and Setgid.
		// XXX(xiaq): Fails.
		/*
			create("su", os.ModeSetuid)
			test("su", true, featureSetuid)
			create("sg", os.ModeSetgid)
			test("sg", true, featureSetgid)
		*/

		// Executable.
		create("xu", 0100)
		create("xg", 0010)
		create("xo", 0001)
		test("xu", true, featureExecutable)
		test("xg", true, featureExecutable)
		test("xo", true, featureExecutable)
	})
}

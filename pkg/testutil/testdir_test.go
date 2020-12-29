package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTestDir_DirIsValid(t *testing.T) {
	dir, cleanup := TestDir()
	defer cleanup()

	stat, err := os.Stat(dir)
	if err != nil {
		t.Errorf("TestDir returns %q which cannot be stated", dir)
	}
	if !stat.IsDir() {
		t.Errorf("TestDir returns %q which is not a dir", dir)
	}
}

func TestTestDir_DirHasSymlinksResolved(t *testing.T) {
	dir, cleanup := TestDir()
	defer cleanup()

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	if dir != resolved {
		t.Errorf("TestDir returns %q, but it resolves to %q", dir, resolved)
	}
}

func TestTestDir_CleanupRemovesDirRecursively(t *testing.T) {
	dir, cleanup := TestDir()

	err := ioutil.WriteFile(filepath.Join(dir, "a"), []byte("test"), 0600)
	if err != nil {
		panic(err)
	}

	cleanup()
	if _, err := os.Stat(dir); err == nil {
		t.Errorf("Dir %q still exists after cleanup", dir)
	}
}

func TestInTestDir_ChangesIntoTempDir(t *testing.T) {
	dir, cleanup := InTestDir()
	defer cleanup()

	pwd := getWd()
	if dir != pwd {
		t.Errorf("InTestDir returns %q but pwd is %q", dir, pwd)
	}
}

func TestInTestDir_CleanupChangesBackToOldWd(t *testing.T) {
	before := getWd()

	_, cleanup := InTestDir()
	cleanup()

	after := getWd()
	if before != after {
		t.Errorf("PWD is %q before InTestDir, but %q after cleanup", before, after)
	}
}

func TestApplyDir_CreatesFiles(t *testing.T) {
	_, cleanup := InTestDir()
	defer cleanup()

	ApplyDir(Dir{
		"a": "a content",
		"b": "b content",
	})

	testFileContent(t, "a", "a content")
	testFileContent(t, "b", "b content")
}

func TestApplyDir_CreatesDirectories(t *testing.T) {
	_, cleanup := InTestDir()
	defer cleanup()

	ApplyDir(Dir{
		"d": Dir{
			"d1": "d1 content",
			"d2": "d2 content",
			"dd": Dir{
				"dd1": "dd1 content",
			},
		},
	})

	testFileContent(t, "d/d1", "d1 content")
	testFileContent(t, "d/d2", "d2 content")
	testFileContent(t, "d/dd/dd1", "dd1 content")
}

func TestApplyDir_Symlinks(t *testing.T) {
	testdir, cleanup := InTestDir()
	defer cleanup()

	ApplyDir(Dir{
		"d1": Dir{
			"f":  "d1/f content",
			"d2": Symlink{testdir},
		},
		"sf1":   Symlink{filepath.Join("d1", "f")},
		"sd1":   Symlink{"d1"},
		"sd2":   Symlink{string(filepath.Separator)},
		"empty": "",
	})

	testFileContent(t, "d1/f", "d1/f content")
	symlinkTargets := [][2]string{
		{"d1/d2", testdir},
		{"sd1", "d1"},
		{"sd2", string(filepath.Separator)},
		{"sf1", filepath.Join("d1", "f")},
	}
	for _, fromTo := range symlinkTargets {
		symlink, expected_target := fromTo[0], fromTo[1]
		fi, err := os.Lstat(symlink)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("File %v isn't a symlink: err %v\n"+
				"os.ModeSymlink %032b\n"+
				"fi.Mode        %032b",
				symlink, err, os.ModeSymlink, fi.Mode())
		}
		actual_target, err := filepath.EvalSymlinks(symlink)
		if err != nil || expected_target != actual_target {
			t.Errorf("Symlink %v is incorrect:\n"+
				"Err: %v\n"+
				"Exp: %v\n"+
				"Act: %v",
				symlink, err, expected_target, actual_target)

		}
	}
}

func getWd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	return dir
}

func testFileContent(t *testing.T, filename string, wantContent string) {
	t.Helper()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Could not read %v: %v", filename, err)
		return
	}
	if string(content) != wantContent {
		t.Errorf("File %v is %q, want %q", filename, content, wantContent)
	}
}

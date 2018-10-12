package util

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

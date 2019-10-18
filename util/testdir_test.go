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

var testDir = Dir{
	"a": "a content",
	"b": "b content",
	"c": "",
	"d": Dir{
		"d1": "d1 content",
		"dd": Dir{
			"dd1": "dd1 content",
		},
	},
}

func TestSetupTestDir_CreatesFiles(t *testing.T) {
	cleanup := SetupTestDir(Dir{
		"a": "a content",
		"b": "b content",
	}, "")
	defer cleanup()

	testFileContent(t, "a", "a content")
	testFileContent(t, "b", "b content")
}

func TestSetupTestDir_CreatesDirectories(t *testing.T) {
	cleanup := SetupTestDir(Dir{
		"d": Dir{
			"d1": "d1 content",
			"d2": "d2 content",
			"dd": Dir{
				"dd1": "dd1 content",
			},
		},
	}, "")
	defer cleanup()

	testFileContent(t, "d/d1", "d1 content")
	testFileContent(t, "d/d2", "d2 content")
	testFileContent(t, "d/dd/dd1", "dd1 content")
}

func TestSetupTestDir_ChangesDirectory(t *testing.T) {
	cleanup := SetupTestDir(Dir{
		"d": Dir{
			"d1": "d1 content",
		},
	}, "d")
	defer cleanup()

	testFileContent(t, "d1", "d1 content")
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

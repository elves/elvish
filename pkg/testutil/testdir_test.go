package testutil

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/tt"
)

func TestTempDir_DirIsValid(t *testing.T) {
	dir := TempDir(t)

	stat, err := os.Stat(dir)
	if err != nil {
		t.Errorf("TestDir returns %q which cannot be stated", dir)
	}
	if !stat.IsDir() {
		t.Errorf("TestDir returns %q which is not a dir", dir)
	}
}

func TestTempDir_DirHasSymlinksResolved(t *testing.T) {
	dir := TempDir(t)

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	if dir != resolved {
		t.Errorf("TestDir returns %q, but it resolves to %q", dir, resolved)
	}
}

func TestTempDir_CleanupRemovesDirRecursively(t *testing.T) {
	c := &cleanuper{}
	dir := TempDir(c)

	err := os.WriteFile(filepath.Join(dir, "a"), []byte("test"), 0600)
	if err != nil {
		panic(err)
	}

	c.runCleanups()
	if _, err := os.Stat(dir); err == nil {
		t.Errorf("Dir %q still exists after cleanup", dir)
	}
}

func TestChdir(t *testing.T) {
	dir := TempDir(t)
	original := getWd()

	c := &cleanuper{}
	Chdir(c, dir)

	after := getWd()
	if after != dir {
		t.Errorf("pwd is now %q, want %q", after, dir)
	}

	c.runCleanups()
	restored := getWd()
	if restored != original {
		t.Errorf("pwd restored to %q, want %q", restored, original)
	}
}

func TestApplyDir_CreatesFiles(t *testing.T) {
	InTempDir(t)

	ApplyDir(Dir{
		"a": "a content",
		"b": "b content",
	})

	testFileContent(t, "a", "a content")
	testFileContent(t, "b", "b content")
}

func TestApplyDir_CreatesDirectories(t *testing.T) {
	InTempDir(t)

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

func TestApplyDir_AllowsExistingDirectories(t *testing.T) {
	InTempDir(t)

	ApplyDir(Dir{"d": Dir{}})
	ApplyDir(Dir{"d": Dir{"a": "content"}})

	testFileContent(t, "d/a", "content")
}

var It = tt.It

func TestDirAsFS(t *testing.T) {
	dir := Dir{
		"d": Dir{
			"x": "this is file d/x",
			"y": "this is file d/y",
		},
		"a": "this is file a",
		"b": File{Perm: 0o777, Content: "this is file b"},
	}

	// fs.WalkDir exercises a large subset of the fs.FS API.
	entries := make(map[string]string)
	fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		must.OK(err)
		entries[path] = fs.FormatFileInfo(must.OK1(d.Info()))
		return nil
	})
	wantEntries := map[string]string{
		".":   "drwxr-xr-x 0 1970-01-01 00:00:00 ./",
		"a":   "-rw-r--r-- 14 1970-01-01 00:00:00 a",
		"b":   "-rwxrwxrwx 14 1970-01-01 00:00:00 b",
		"d":   "drwxr-xr-x 0 1970-01-01 00:00:00 d/",
		"d/x": "-rw-r--r-- 16 1970-01-01 00:00:00 x",
		"d/y": "-rw-r--r-- 16 1970-01-01 00:00:00 y",
	}
	if diff := cmp.Diff(wantEntries, entries); diff != "" {
		t.Errorf("DirEntry map from walking the FS: (-want +got):\n%s", diff)
	}

	// Direct file access is not exercised by fs.WalkDir (other than to "."), so
	// test those too.
	readFile := func(name string) (string, error) {
		bs, err := fs.ReadFile(dir, name)
		return string(bs), err
	}
	tt.Test(t, tt.Fn(readFile).Named("readFile"),
		It("supports accessing file in root").
			Args("a").
			Rets("this is file a", error(nil)),
		It("supports accessing file backed by a File struct").
			Args("b").
			Rets("this is file b", error(nil)),
		It("supports accessing file in subdirectory").
			Args("d/x").
			Rets("this is file d/x", error(nil)),
		It("errors if file doesn't exist").
			Args("d/bad").
			Rets("", &fs.PathError{Op: "open", Path: "d/bad", Err: fs.ErrNotExist}),
		It("errors if a directory component of the path doesn't exist").
			Args("badd/x").
			Rets("", &fs.PathError{Op: "open", Path: "badd/x", Err: fs.ErrNotExist}),
		It("errors if a directory component of the path is a file").
			Args("a/x").
			Rets("", &fs.PathError{Op: "open", Path: "a/x", Err: fs.ErrNotExist}),
		It("can open but not read a directory").
			Args("d").
			Rets("", &fs.PathError{Op: "read", Path: "d", Err: errIsDir}),
		It("errors if path is invalid").
			Args("/d").
			Rets("", &fs.PathError{Op: "open", Path: "/d", Err: fs.ErrInvalid}),
	)

	// fs.WalkDir calls ReadDir with -1. Also exercise the code for reading
	// piece by piece.
	file := must.OK1(dir.Open(".")).(fs.ReadDirFile)
	rootEntries := make(map[string]string)
	for {
		es, err := file.ReadDir(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		rootEntries[es[0].Name()] = fs.FormatFileInfo(must.OK1(es[0].Info()))
	}
	wantRootEntries := map[string]string{
		"a": "-rw-r--r-- 14 1970-01-01 00:00:00 a",
		"b": "-rwxrwxrwx 14 1970-01-01 00:00:00 b",
		"d": "drwxr-xr-x 0 1970-01-01 00:00:00 d/",
	}
	if diff := cmp.Diff(wantRootEntries, rootEntries); diff != "" {
		t.Errorf("DirEntry map from reading the root piece by piece: (-want +got):\n%s", diff)
	}

	// Cover the Sys method of the two FileInfo implementations.
	must.OK1(must.OK1(dir.Open("d")).Stat()).Sys()
	must.OK1(must.OK1(dir.Open("a")).Stat()).Sys()
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
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Could not read %v: %v", filename, err)
		return
	}
	if string(content) != wantContent {
		t.Errorf("File %v is %q, want %q", filename, content, wantContent)
	}
}

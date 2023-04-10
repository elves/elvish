package glob

import (
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"src.elv.sh/pkg/testutil"
)

var (
	mkdirs = []string{"a", "b", "c", "d1", "d1/e", "d1/e/f", "d1/e/f/g",
		"d2", "d2/e", "d2/e/f", "d2/e/f/g"}
	mkdirDots = []string{".el"}
	creates   = []string{"a/X", "a/Y", "b/X", "c/Y",
		"dX", "dXY",
		"lorem", "ipsum",
		"d1/e/f/g/X", "d2/e/f/g/X"}
	createDots = []string{".x", ".el/x"}
	symlinks   = []struct {
		path   string
		target string
	}{
		{"d1/s-f", "f"},
		{"s-d", "d2"},
		{"s-d-f", "d1/f"},
		{"s-bad", "bad"},
	}
)

type globCase struct {
	pattern string
	want    []string
}

var globCases = []globCase{
	{"*", []string{"a", "b", "c", "d1", "d2", "dX", "dXY", "lorem", "ipsum", "s-bad", "s-d", "s-d-f"}},
	{".", []string{"."}},
	{"./*", []string{"./a", "./b", "./c", "./d1", "./d2", "./dX", "./dXY", "./lorem", "./ipsum", "./s-bad", "./s-d", "./s-d-f"}},
	{"..", []string{".."}},
	{"a/..", []string{"a/.."}},
	{"a/../*", []string{"a/../a", "a/../b", "a/../c", "a/../d1", "a/../d2", "a/../dX", "a/../dXY", "a/../ipsum", "a/../lorem", "a/../s-bad", "a/../s-d", "a/../s-d-f"}},
	{"*/", []string{"a/", "b/", "c/", "d1/", "d2/"}},
	{"**", []string{"a", "a/X", "a/Y", "b", "b/X", "c", "c/Y", "d1", "d1/e", "d1/e/f", "d1/e/f/g", "d1/e/f/g/X", "d1/s-f", "d2", "d2/e", "d2/e/f", "d2/e/f/g", "d2/e/f/g/X", "dX", "dXY", "ipsum", "lorem", "s-bad", "s-d", "s-d-f"}},
	{"*/X", []string{"a/X", "b/X"}},
	{"**X", []string{"a/X", "b/X", "dX", "d1/e/f/g/X", "d2/e/f/g/X"}},
	{"*/*/*", []string{"d1/e/f", "d2/e/f"}},
	{"l*m", []string{"lorem"}},
	{"d*", []string{"d1", "d2", "dX", "dXY"}},
	{"d*/", []string{"d1/", "d2/"}},
	{"d**", []string{"d1", "d1/e", "d1/e/f", "d1/e/f/g", "d1/e/f/g/X", "d1/s-f", "d2", "d2/e", "d2/e/f", "d2/e/f/g", "d2/e/f/g/X", "dX", "dXY"}},
	{"?", []string{"a", "b", "c"}},
	{"??", []string{"d1", "d2", "dX"}},

	// Nonexistent paths.
	{"xxxx", []string{}},
	{"xxxx/*", []string{}},
	{"a/*/", []string{}},

	// TODO: Add more tests for situations where Lstat fails.

	// TODO Test cases against dotfiles.
}

func TestGlob_Relative(t *testing.T) {
	testGlob(t, false)
}

func TestGlob_Absolute(t *testing.T) {
	testGlob(t, true)
}

func testGlob(t *testing.T, abs bool) {
	dir := testutil.InTempDir(t)
	dir = strings.ReplaceAll(dir, string(os.PathSeparator), "/")

	for _, dir := range append(mkdirs, mkdirDots...) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
	for _, file := range append(creates, createDots...) {
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
	for _, link := range symlinks {
		err := os.Symlink(link.target, link.path)
		if err != nil {
			// Creating symlinks requires a special permission on Windows. If
			// the user doesn't have that permission, create the symlink as an
			// ordinary file instead.
			f, err := os.Create(link.path)
			if err != nil {
				panic(err)
			}
			f.Close()
		}
	}

	for _, tc := range globCases {
		pattern := tc.pattern
		if abs {
			pattern = dir + "/" + pattern
		}
		wantResults := make([]string, len(tc.want))
		for i, result := range tc.want {
			if abs {
				wantResults[i] = dir + "/" + result
			} else {
				wantResults[i] = result
			}
		}
		sort.Strings(wantResults)

		results := globPaths(pattern)

		if !reflect.DeepEqual(results, wantResults) {
			t.Errorf(`Glob(%q) => %v, want %v`, pattern, results, wantResults)
		}
	}
}

// Regression test for b.elv.sh/1220
func TestGlob_InvalidUTF8InFilename(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows, filenames are converted to UTF-16 before being passed
		// to API calls, meaning that all the invalid byte sequences will be
		// normalized to U+FFFD, making this impossible to test.
		t.Skip()
	}

	testutil.InTempDir(t)

	name := string([]byte{255}) + "x"
	f, err := os.Create(name)
	if err != nil {
		// The system may refuse to create a file whose name is not UTF-8. This
		// happens on macOS 11 with an APFS filesystem.
		t.Skip("create: ", err)
	}
	f.Close()

	paths := globPaths("*x")
	wantPaths := []string{name}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Errorf("got %v, want %v", paths, wantPaths)
	}
}

func globPaths(pattern string) []string {
	paths := []string{}
	Glob(pattern, func(pathInfo PathInfo) bool {
		paths = append(paths, pathInfo.Path)
		return true
	})
	sort.Strings(paths)
	return paths
}

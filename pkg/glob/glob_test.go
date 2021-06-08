package glob

import (
	"os"
	"reflect"
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
)

type globCase struct {
	pattern string
	want    []string
}

var globCases = []globCase{
	{"*", []string{"a", "b", "c", "d1", "d2", "dX", "dXY", "lorem", "ipsum"}},
	{".", []string{"."}},
	{"./*", []string{"./a", "./b", "./c", "./d1", "./d2", "./dX", "./dXY", "./lorem", "./ipsum"}},
	{"..", []string{".."}},
	{"a/..", []string{"a/.."}},
	{"a/../*", []string{"a/../a", "a/../b", "a/../c", "a/../d1", "a/../d2", "a/../dX", "a/../dXY", "a/../lorem", "a/../ipsum"}},
	{"*/", []string{"a/", "b/", "c/", "d1/", "d2/"}},
	{"**", append(mkdirs, creates...)},
	{"*/X", []string{"a/X", "b/X"}},
	{"**X", []string{"a/X", "b/X", "dX", "d1/e/f/g/X", "d2/e/f/g/X"}},
	{"*/*/*", []string{"d1/e/f", "d2/e/f"}},
	{"l*m", []string{"lorem"}},
	{"d*", []string{"d1", "d2", "dX", "dXY"}},
	{"d*/", []string{"d1/", "d2/"}},
	{"d**", []string{"d1", "d1/e", "d1/e/f", "d1/e/f/g", "d1/e/f/g/X",
		"d2", "d2/e", "d2/e/f", "d2/e/f/g", "d2/e/f/g/X", "dX", "dXY"}},
	{"?", []string{"a", "b", "c"}},
	{"??", []string{"d1", "d2", "dX"}},

	// Nonexistent paths.
	{"xxxx", []string{}},
	{"xxxx/*", []string{}},
	{"a/*/", []string{}},

	// TODO Test cases against dotfiles.
}

func TestGlob_Relative(t *testing.T) {
	testGlob(t, false)
}

func TestGlob_Absolute(t *testing.T) {
	testGlob(t, true)
}

func testGlob(t *testing.T, abs bool) {
	dir, cleanup := testutil.InTestDir()
	dir = strings.ReplaceAll(dir, string(os.PathSeparator), "/")
	defer cleanup()

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
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	name := string([]byte{255}) + ".c"
	f, err := os.Create(name)
	if err != nil {
		// The system may refuse to create a file whose name is not UTF-8. This
		// happens on macOS 11 with an APFS filesystem.
		t.Skip("create: ", err)
	}
	f.Close()

	_, err = os.Stat(name)
	if err != nil {
		// The system may pretend to have created the file successfully,
		// but substitute the invalid sequences with U+FFFD. This happens on
		// Windows 10 with an NTFS filesystem.
		t.Skip("stat: ", err)
	}

	paths := globPaths("*.c")
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

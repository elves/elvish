package glob

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/util"
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
	dir, cleanup := util.InTestDir()
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

		results := []string{}
		Glob(pattern, func(name string) bool {
			results = append(results, name)
			return true
		})
		sort.Strings(results)

		if !reflect.DeepEqual(results, wantResults) {
			t.Errorf(`Glob(%q) => %v, want %v`, pattern, results, wantResults)
		}
	}
}

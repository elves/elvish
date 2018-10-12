package glob

import (
	"os"
	"reflect"
	"runtime"
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

func init() {
	// Add tests for absolute paths. This is platform-dependent and may break if
	// the directories used change during testing.
	var dirs []string
	if runtime.GOOS == "windows" {
		dirs = []string{`C:/`, `C:/Windows/`}
	} else {
		dirs = []string{"/", "/usr/"}
	}
	for _, dir := range dirs {
		globCases = append(globCases, globCase{dir + "*", util.FullNames(dir)})
	}
}

func TestGlob(t *testing.T) {
	_, cleanup := util.InTestDir()
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
		names := []string{}
		Glob(tc.pattern, func(name string) bool {
			names = append(names, name)
			return true
		})
		sort.Strings(names)
		sort.Strings(tc.want)
		if !reflect.DeepEqual(names, tc.want) {
			t.Errorf(`Glob(%q, "") => %v, want %v`, tc.pattern, names, tc.want)
		}
	}
}

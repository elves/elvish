package glob

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
)

var (
	mkdirs = []string{"a", "b", "c", "d1", "d1/e", "d1/e/f", "d1/e/f/g",
		"d2", "d2/e", "d2/e/f", "d2/e/f/g"}
	creates = []string{"a/X", "a/Y", "b/X", "c/Y", "dX", "lorem", "ipsum",
		"d1/e/f/g/X", "d2/e/f/g/X"}
)

var globCases = []struct {
	pattern string
	want    []string
}{
	{"*", []string{"a", "b", "c", "d1", "d2", "dX", "lorem", "ipsum"}},
	{"*/", []string{"a/", "b/", "c/", "d1/", "d2/"}},
	{"**", append(mkdirs, creates...)},
	{"*/X", []string{"a/X", "b/X"}},
	{"**X", []string{"a/X", "b/X", "dX", "d1/e/f/g/X", "d2/e/f/g/X"}},
	{"*/*/*", []string{"d1/e/f", "d2/e/f"}},
	{"l*m", []string{"lorem"}},
	{"d*", []string{"d1", "d2", "dX"}},
	{"d*/", []string{"d1/", "d2/"}},
	{"d**", []string{"d1", "d1/e", "d1/e/f", "d1/e/f/g", "d1/e/f/g/X",
		"d2", "d2/e", "d2/e/f", "d2/e/f/g", "d2/e/f/g/X", "dX"}},
}

func TestGlob(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "glob-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)
	os.Chdir(tmpdir)

	for _, dir := range mkdirs {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
	for _, file := range creates {
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
	for _, tc := range globCases {
		names := Glob(tc.pattern)
		sort.Strings(names)
		sort.Strings(tc.want)
		if !reflect.DeepEqual(names, tc.want) {
			t.Errorf(`Glob(%q, "") => %v, want %v`, tc.pattern, names, tc.want)
		}
	}
}

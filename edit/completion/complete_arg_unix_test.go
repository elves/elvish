// +build !windows,!plan9

// TODO: Enable on Windows.
package completion

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

var (
	fileStyle = ui.StylesFromString("1")
	exeStyle  = ui.StylesFromString("2")
	dirStyle  = ui.StylesFromString("4")
)

var complFilenameInnerTests = []struct {
	head           string
	executableOnly bool
	wantCandidates rawCandidates
}{
	// Match all non-hidden files and dirs, in alphabetical order.
	// Files have suffix " " and directories "/". Styles are set according to
	// the LS_COLORS variable, which are set in the beginning of the test.
	{"haha", false, rawCandidates{
		&complexCandidate{stem: "Documents", codeSuffix: "/", style: dirStyle},
		&complexCandidate{stem: "bar", codeSuffix: " ", style: fileStyle},
		&complexCandidate{stem: "elvish", codeSuffix: " ", style: exeStyle},
		&complexCandidate{stem: "foo", codeSuffix: " ", style: fileStyle},
	}},
	// Only match executables and directories.
	{"haha", true, rawCandidates{
		&complexCandidate{stem: "Documents", codeSuffix: "/", style: dirStyle},
		&complexCandidate{stem: "elvish", codeSuffix: " ", style: exeStyle},
	}},
	// Match hidden files and directories.
	{".haha", false, rawCandidates{
		&complexCandidate{stem: ".elvish", codeSuffix: "/", style: dirStyle},
		&complexCandidate{stem: ".vimrc", codeSuffix: " ", style: fileStyle},
	}},
}

func TestComplFilenameInner(t *testing.T) {
	os.Setenv("LS_COLORS", "rs=1:ex=2:di=4")
	_, cleanup := util.InTestDir()
	defer cleanup()

	create("foo", 0600)
	create(".vimrc", 0600)
	create("bar", 0600)

	create("elvish", 0700)

	mkdir("Documents", 0700)
	mkdir(".elvish", 0700)

	for _, test := range complFilenameInnerTests {
		var (
			err   error
			cands rawCandidates
			gets  = make(chan rawCandidate)
		)
		go func() {
			defer close(gets)
			err = complFilenameInner(test.head, test.executableOnly, gets)
		}()
		for v := range gets {
			cands = append(cands, v)
		}
		if err != nil {
			t.Errorf("complFilenameInner(%v, %v) returns error %v, want nil",
				test.head, test.executableOnly, err)
		}
		sort.Sort(cands)
		if !reflect.DeepEqual(cands, test.wantCandidates) {
			t.Errorf("complFilenameInner(%v, %v) returns %v, want %v",
				test.head, test.executableOnly, cands, test.wantCandidates)
			t.Log("returned candidates are:")
			for _, cand := range cands {
				t.Logf("%#v", cand)
			}
		}
	}
}

func mkdir(dirname string, perm os.FileMode) {
	err := os.Mkdir(dirname, perm)
	if err != nil {
		panic(err)
	}
}

func create(fname string, perm os.FileMode) {
	f, err := os.OpenFile(fname, os.O_CREATE, perm)
	if err != nil {
		panic(err)
	}
	f.Close()
}

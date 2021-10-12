package fsutil

import (
	"reflect"
	"sort"
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestEachExternal(t *testing.T) {
	binPath := testutil.InTempDir(t)

	testutil.Setenv(t, "PATH", "Z:\\foo;"+binPath+";Z:\\bar")

	testutil.ApplyDir(testutil.Dir{
		"dir":      testutil.Dir{},
		"file.txt": "",
		"cmd.bat":  testutil.File{Perm: 0755, Content: ""},
		"cmd.cmd":  testutil.File{Perm: 0755, Content: ""},
		"cmd.exe":  "",
		"cmd.txt":  testutil.File{Perm: 0755, Content: ""},
	})

	wantCmds := []string{"cmd.bat", "cmd.cmd", "cmd.exe"}
	gotCmds := []string{}
	EachExternal(func(cmd string) { gotCmds = append(gotCmds, cmd) })

	sort.Strings(gotCmds)
	if !reflect.DeepEqual(wantCmds, gotCmds) {
		t.Errorf("EachExternal want %q got %q", wantCmds, gotCmds)
	}
}

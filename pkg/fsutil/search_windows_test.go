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
		"prog.bat": testutil.File{Perm: 0o666, Content: ""},
		"prog.cmd": testutil.File{Perm: 0o755, Content: ""},
		"prog.txt": testutil.File{Perm: 0o755, Content: ""},
		"PROG.EXE": "", // validate that explicit file perms don't matter
	})

	wantCmds := []string{"prog.bat", "prog.cmd", "PROG.EXE"}
	gotCmds := []string{}
	EachExternal(func(cmd string) { gotCmds = append(gotCmds, cmd) })

	sort.Strings(wantCmds)
	sort.Strings(gotCmds)
	if !reflect.DeepEqual(wantCmds, gotCmds) {
		t.Errorf("EachExternal want %q got %q", wantCmds, gotCmds)
	}
}

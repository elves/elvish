package fsutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

	if diff := cmp.Diff(wantCmds, gotCmds, sortStringSlices); diff != "" {
		t.Errorf("EachExternal (-want +got): \n%s", diff)
	}
}

var sortStringSlices = cmpopts.SortSlices(func(a, b string) bool { return a < b })

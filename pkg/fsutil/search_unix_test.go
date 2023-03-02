//go:build unix

package fsutil

import (
	"reflect"
	"sort"
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestEachExternal(t *testing.T) {
	binPath := testutil.InTempDir(t)

	testutil.Setenv(t, "PATH", "/foo:"+binPath+":/bar")

	testutil.ApplyDir(testutil.Dir{
		"dir":  testutil.Dir{},
		"file": "",
		"cmdx": "#!/bin/sh",
		"cmd1": testutil.File{Perm: 0755, Content: "#!/bin/sh"},
		"cmd2": testutil.File{Perm: 0755, Content: "#!/bin/sh"},
		"cmd3": testutil.File{Perm: 0755, Content: ""},
	})

	wantCmds := []string{"cmd1", "cmd2", "cmd3"}
	gotCmds := []string{}
	EachExternal(func(cmd string) { gotCmds = append(gotCmds, cmd) })

	sort.Strings(gotCmds)
	if !reflect.DeepEqual(wantCmds, gotCmds) {
		t.Errorf("EachExternal want %q got %q", wantCmds, gotCmds)
	}
}

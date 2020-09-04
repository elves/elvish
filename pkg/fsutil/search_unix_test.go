// +build !windows,!plan9,!js

package fsutil

import (
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/pkg/testutil"
)

// TODO: When EachExternal is modified to work on Windows either fold this
// test into external_cmd_test.go or create an external_cmd_windows_test.go
// that performs an equivalent test on Windows.
func TestEachExternal(t *testing.T) {
	binPath, cleanup := testutil.InTestDir()
	defer cleanup()

	restorePath := testutil.WithTempEnv("PATH", "/foo:"+binPath+":/bar")
	defer restorePath()

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

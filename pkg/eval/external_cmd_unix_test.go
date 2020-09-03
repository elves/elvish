// +build !windows,!plan9,!js

package eval_test

import (
	"reflect"
	"sort"
	"syscall"
	"testing"

	. "github.com/elves/elvish/pkg/eval"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/testutil"
)

// TODO: When EachExternal is modified to work on Windows either fold this
// test into external_cmd_test.go or create an external_cmd_windows_test.go
// that performs an equivalent test on Windows.
func TestEachExternal(t *testing.T) {
	tmpHome, cleanup := InTempHome()
	defer cleanup()

	restorePath := testutil.WithTempEnv("PATH", "/foo:"+tmpHome+":/bar")
	defer restorePath()

	MustMkdirAll("dir")
	MustCreateEmpty("cmdx")
	MustWriteFile("cmd1", []byte("#!/bin/sh"), 0755)
	MustWriteFile("cmd2", []byte("#!/bin/sh"), 0755)
	MustWriteFile("cmd3", []byte(""), 0755)
	MustCreateEmpty("file")

	wantCmds := []string{"cmd1", "cmd2", "cmd3"}
	gotCmds := []string{}
	EachExternal(func(filename string) {
		gotCmds = append(gotCmds, filename)
	})

	sort.Strings(gotCmds)
	if !reflect.DeepEqual(wantCmds, gotCmds) {
		t.Errorf("EachExternal want %q got %q", wantCmds, gotCmds)
	}
}

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	// The exit<<8 is gross but I can't find any exported symbols that would
	// allow us to construct WaitStatus. So assume legacy UNIX encoding
	// for a process that exits normally; i.e., not due to a signal.
	return syscall.WaitStatus(exit << 8)
}

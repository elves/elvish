// +build !windows,!plan9,!js

package eval

import (
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

// TODO: When EachExternal is modified to work on Windows either fold this
// test into external_cmd_test.go or create an external_cmd_windows_test.go
// that performs an equivalent test on Windows.
func TestEachExternal(t *testing.T) {
	tmpHome, cleanup := InTempHome()
	defer cleanup()

	restorePath := util.WithTempEnv("PATH", "/foo:"+tmpHome+":/bar")
	defer restorePath()

	mustMkdirAll("dir")
	mustCreateEmpty("cmdx")
	mustWriteFile("cmd1", []byte("#!/bin/sh"), 0755)
	mustWriteFile("cmd2", []byte("#!/bin/sh"), 0755)
	mustWriteFile("cmd3", []byte(""), 0755)
	mustCreateEmpty("file")

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

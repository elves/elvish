package eval

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestBuiltinFnExternal(t *testing.T) {
	tmpHome, cleanup := InTempHome()
	defer cleanup()

	mustMkdirAll("dir")
	mustWriteFile("cmd1", []byte("#!/bin/sh"), 0755)
	mustWriteFile("cmd2", []byte("#!/bin/sh"), 0755)
	mustWriteFile("cmd3", []byte(""), 0755)
	mustCreateEmpty("cmdx")
	mustCreateEmpty("file")

	os.Setenv("PATH", tmpHome+os.Getenv("PATH"))
	Test(t,
		That(`resolve external`).Puts("$external~"),
		That(`e = (external true); kind-of $e`).Puts("external"),
		That(`e = (external true); put (repr $e)`).Puts("<external true>"),
		That(`e = (external false); m = [&$e=true]; put (repr $m)`).Puts("[&<external false>=true]"),
		// This group tests the `ExternalCmd.Call` method.
		That(`e = (external true); $e`).DoesNothing(),
		That(`e = (external true); $e &option`).Throws(ErrExternalCmdOpts, "$e &option"),
		// The 1<<8 is gross but I can't find any exported symbols that would
		// allow us to construct WaitStatus. So assume legacy UNIX encoding
		// for a process that exits with status one.
		That(`e = (external false); $e`).ThrowsCmdExit(NewExternalCmdExit("false", 1<<8, 0)),
		// The command shouldn't be found when run so we should get an
		// exception along the lines of "executable file not found in $PATH".
		That(`e = (external true); E:PATH=/; $e`).ThrowsAny(),
		// The command will be found but cause an exception along the lines of
		// "exec format error" because it is marked executable but is empty.
		That(`e = (external cmd3); E:PATH=/; $e`).ThrowsAny(),
	)

	os.Setenv("PATH", "/argle:"+tmpHome+":/bargle")
	want_cmds := []string{"cmd1", "cmd2", "cmd3"}
	got_cmds := []string{}
	EachExternal(func(filename string) {
		got_cmds = append(got_cmds, filename)
	})

	sort.Strings(got_cmds)
	if !reflect.DeepEqual(want_cmds, got_cmds) {
		t.Errorf("EachExternal want %q got %q", want_cmds, got_cmds)
	}
}

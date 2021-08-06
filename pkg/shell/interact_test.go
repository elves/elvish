package shell

import (
	"path/filepath"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
	. "src.elv.sh/pkg/testutil"
)

func TestInteract_SingleCommand(t *testing.T) {
	f := setup(t)
	f.FeedIn("echo hello\n")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello\n")
}

func TestInteract_Exception(t *testing.T) {
	f := setup(t)
	f.FeedIn("fail mock\n")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_LegacyRcFile(t *testing.T) {
	f := setup(t)
	MustWriteFile(
		filepath.Join(f.home, ".elvish", "rc.elv"), "echo hello legacy rc.elv")
	f.FeedIn("")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello legacy rc.elv\n")
}

// Non-legacy RC file tested in interact_unix_test.go

func TestInteract_RcFile(t *testing.T) {
	f := setup(t)
	f.FeedIn("")
	MustWriteFile("rc.elv", "echo hello from rc.elv")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := setup(t)
	f.FeedIn("")
	MustWriteFile("rc.elv", "echo $a")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := setup(t)
	f.FeedIn("")
	MustWriteFile("rc.elv", "fail mock")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_NonexistentIsOK(t *testing.T) {
	f := setup(t)
	f.FeedIn("")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "")
}

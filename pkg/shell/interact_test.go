package shell

import (
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestInteract_SingleCommand(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("echo hello\n")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello\n")
}

func TestInteract_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("fail mock\n")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	MustWriteFile("rc.elv", "echo hello from rc.elv")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	MustWriteFile("rc.elv", "echo $a")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	MustWriteFile("rc.elv", "fail mock")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_NonexistentIsOK(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "")
}

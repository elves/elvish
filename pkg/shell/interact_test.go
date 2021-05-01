package shell

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestMain(m *testing.M) {
	interactiveRescueShell = false
	os.Exit(m.Run())
}

func TestInteract_SingleCommand(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("echo hello\n")

	Interact(f.Fds(), &InteractConfig{})
	f.TestOut(t, 1, "hello\n")
}

func TestInteract_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("fail mock\n")

	Interact(f.Fds(), &InteractConfig{})
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo hello from rc.elv")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo $a")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "fail mock")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_NonexistentIsOK(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOut(t, 1, "")
}

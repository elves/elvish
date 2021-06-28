package shell

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/prog/progtest"
)

func init() {
	interactiveRescueShell = false
}

func TestInteract_SingleCommand(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("echo hello\n")

	Interact(f.Fds(), interactConfig(""))
	f.TestOut(t, 1, "hello\n")
}

func TestInteract_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("fail mock\n")

	Interact(f.Fds(), interactConfig(""))
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo hello from rc.elv")

	Interact(f.Fds(), interactConfig("rc.elv"))
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo $a")

	Interact(f.Fds(), interactConfig("rc.elv"))
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "fail mock")

	Interact(f.Fds(), interactConfig("rc.elv"))
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_NonexistentIsOK(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	Interact(f.Fds(), interactConfig("rc.elv"))
	f.TestOut(t, 1, "")
}

func interactConfig(rc string) *InteractConfig {
	return &InteractConfig{Evaler: eval.NewEvaler(), RC: rc}
}

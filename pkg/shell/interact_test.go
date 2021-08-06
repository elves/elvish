package shell

import (
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
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

func TestInteract_LegacyRcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	home, cleanupHome := tempDirEnv(env.HOME)
	defer cleanupHome()
	testutil.MustWriteFile(
		filepath.Join(home, ".elvish", "rc.elv"),
		"echo hello legacy rc.elv")
	f.FeedIn("")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello legacy rc.elv\n")
}

func TestInteract_NewRcFile_Default(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	home, cleanupHome := tempDirEnv(env.HOME)
	defer cleanupHome()
	testutil.MustWriteFile(
		filepath.Join(home, ".config", "elvish", "rc.elv"),
		"echo hello new rc.elv")

	f.FeedIn("")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello new rc.elv\n")
}

func TestInteract_NewRcFile_XDG_CONFIG_HOME(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	// Make an empty $HOME to prevent real ~/.elvish/rc.elv from being used
	_, cleanupHome := tempDirEnv(env.HOME)
	defer cleanupHome()
	xdgConfigHome, cleanupDir := tempDirEnv(env.XDG_CONFIG_HOME)
	defer cleanupDir()
	testutil.MustWriteFile(
		filepath.Join(xdgConfigHome, "elvish", "rc.elv"),
		"echo hello XDG_CONFIG_HOME rc.elv")

	f.FeedIn("")

	exit := run(f.Fds(), Elvish())
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello XDG_CONFIG_HOME rc.elv\n")
}

func TestInteract_RcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	testutil.MustWriteFile("rc.elv", "echo hello from rc.elv")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	testutil.MustWriteFile("rc.elv", "echo $a")

	exit := run(f.Fds(), Elvish("-rc", "rc.elv"))
	TestExit(t, exit, 0)
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")
	testutil.MustWriteFile("rc.elv", "fail mock")

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

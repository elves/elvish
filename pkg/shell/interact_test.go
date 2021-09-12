package shell

import (
	"path/filepath"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
	. "src.elv.sh/pkg/testutil"
)

func TestInteract(t *testing.T) {
	setupHomePaths(t)
	InTempDir(t)
	MustWriteFile("rc.elv", "echo hello from rc.elv")
	MustWriteFile("rc-dnc.elv", "echo $a")
	MustWriteFile("rc-fail.elv", "fail bad")

	Test(t, Program{},
		thatElvishInteract().WithStdin("echo hello\n").WritesStdout("hello\n"),
		thatElvishInteract().WithStdin("fail mock\n").WritesStderrContaining("fail mock"),

		thatElvishInteract("-rc", "rc.elv").WritesStdout("hello from rc.elv\n"),
		// rc file does not compile
		thatElvishInteract("-rc", "rc-dnc.elv").
			WritesStderrContaining("variable $a not found"),
		// rc file throws exception
		thatElvishInteract("-rc", "rc-fail.elv").WritesStderrContaining("fail bad"),
		// rc file not existing is OK
		thatElvishInteract("-rc", "rc-nonexistent.elv").DoesNothing(),
	)
}

func TestInteract_DefaultRCPath(t *testing.T) {
	home := setupHomePaths(t)
	// Legacy RC path
	MustWriteFile(
		filepath.Join(home, ".elvish", "rc.elv"), "echo hello legacy rc.elv")
	// Note: non-legacy path is tested in interact_unix_test.go

	Test(t, Program{},
		thatElvishInteract().WritesStdout("hello legacy rc.elv\n"),
	)
}

func thatElvishInteract(args ...string) Case {
	return ThatElvish(args...).WritesStderrContaining("")
}

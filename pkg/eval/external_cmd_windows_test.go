//go:build windows

package eval_test

import (
	"syscall"
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestExternalCmd_Windows_RelativePathWithSlashes(t *testing.T) {
	testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"foo.bat": "@echo foo",
		"dir": testutil.Dir{
			"bar.bat": "@echo bar",
			"baz": testutil.Dir{
				"lorem.bat": "@echo lorem",
			},
		},
	})
	testutil.Chdir(t, "dir")

	Test(t,
		That("../foo").Prints("foo\r\n"),
		That("./bar").Prints("bar\r\n"),
		That("./baz/lorem").Prints("lorem\r\n"),
		That("baz/lorem").Prints("lorem\r\n"),
	)
}

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	return syscall.WaitStatus{ExitCode: exit}
}

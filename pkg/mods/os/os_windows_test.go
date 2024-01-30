package os_test

import (
	"testing"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func mkFifoOrSkip(t *testing.T, _ string) {
	t.Skip("can't make FIFO on Windows")
}

func createWindowsSpecialFileOrSkip(t *testing.T) {
	testutil.ApplyDir(testutil.Dir{
		"directory": testutil.Dir{},
		"readonly":  "",
		"hidden":    "",
	})
	mustSetFileAttributes("readonly", windows.FILE_ATTRIBUTE_READONLY)
	mustSetFileAttributes("hidden", windows.FILE_ATTRIBUTE_HIDDEN)
}

func mustSetFileAttributes(name string, attr uint32) {
	must.OK(windows.SetFileAttributes(must.OK1(windows.UTF16PtrFromString(name)), attr))
}

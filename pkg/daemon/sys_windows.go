package daemon

import (
	"os"
	"syscall"
)

// https://docs.microsoft.com/en-us/windows/win32/winsock/windows-sockets-error-codes-2
var errConnRefused = syscall.Errno(10061)

// No-op on Windows.
func setUmaskForDaemon() {}

// A subset of possible process creation flags, value taken from
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
const (
	createBreakwayFromJob = 0x01000000
	createNewProcessGroup = 0x00000200
	detachedProcess       = 0x00000008
	daemonCreationFlags   = createBreakwayFromJob | createNewProcessGroup | detachedProcess
)

func procAttrForSpawn(files []*os.File) *os.ProcAttr {
	return &os.ProcAttr{
		Dir:   `C:\`,
		Env:   []string{"SystemRoot=" + os.Getenv("SystemRoot")}, // SystemRoot is needed for net.Listen for some reason
		Files: files,
		Sys:   &syscall.SysProcAttr{CreationFlags: daemonCreationFlags},
	}
}

package daemon

import (
	"os"
	"syscall"
)

// A subset of possible process creation flags, value taken from
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
const (
	CREATE_BREAKAWAY_FROM_JOB = 0x01000000
	CREATE_NEW_PROCESS_GROUP  = 0x00000200
	DETACHED_PROCESS          = 0x00000008

	daemonCreationFlags = CREATE_BREAKAWAY_FROM_JOB | CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS
)

func procAttrForSpawn(stdout *os.File) *os.ProcAttr {
	// The daemon should not inherit a reference to the stdin file descriptor
	// of the original shell. The daemon shouldn't read anything from stdin so
	// use the null device.
	devnull, _ := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	return &os.ProcAttr{
		Dir:   `C:\`,
		Env:   []string{"SystemRoot=" + os.Getenv("SystemRoot")}, // SystemRoot is needed for net.Listen for some reason
		Files: []*os.File{devnull, stdout, stdout},
		Sys:   &syscall.SysProcAttr{CreationFlags: daemonCreationFlags},
	}
}

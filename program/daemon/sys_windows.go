package daemon

import (
	"os"
	"syscall"
)

func setUmask() {
	// NOP on windows.
}

// A subset of possible process creation flags, value taken from
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
const (
	CREATE_BREAKAWAY_FROM_JOB = 0x01000000
	CREATE_NEW_PROCESS_GROUP  = 0x00000200
	DETACHED_PROCESS          = 0x00000008

	DaemonCreationFlags = CREATE_BREAKAWAY_FROM_JOB | CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS
)

func procAttrForSpawn() *os.ProcAttr {
	return &os.ProcAttr{
		Dir:   `C:\`,
		Env:   []string{"SystemRoot=" + os.Getenv("SystemRoot")}, // SystemRoot is needed for net.Listen for some reason
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   &syscall.SysProcAttr{CreationFlags: DaemonCreationFlags},
	}
}

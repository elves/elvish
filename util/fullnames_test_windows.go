package util

import (
	"os/exec"
	"sort"
	"strings"
	"syscall"
)

func ls(dir string) []string {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: "cmd /C dir /A /B " + dir,
	}
	output, err := cmd.Output()
	mustOK(err)
	names := strings.Split(strings.Trim(string(output), "\r\n"), "\r\n")
	for i := range names {
		names[i] = dir + names[i]
	}
	sort.Strings(names)
	return names
}

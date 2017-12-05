package util

import (
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"testing"
)

func ls(dir string) []string {
	var names []string
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CmdLine: "cmd /C dir /A /B " + dir,
		}
		output, err := cmd.Output()
		mustOK(err)
		names = strings.Split(strings.Trim(string(output), "\r\n"), "\r\n")
	} else {
		// BUG: will fail if there are filenames containing newlines.
		output, err := exec.Command("ls", dir).Output()
		mustOK(err)
		names = strings.Split(strings.Trim(string(output), "\n"), "\n")
	}
	for i := range names {
		names[i] = dir + names[i]
	}
	sort.Strings(names)
	return names
}

func TestFullNames(t *testing.T) {
	var dirs []string
	if runtime.GOOS == "windows" {
		dirs = []string{`C:\`, `C:\Users\`}
	} else {
		dirs = []string{"/", "/usr"}
	}
	for _, dir := range dirs {
		wantNames := ls(dir)
		names := FullNames(dir)
		if !reflect.DeepEqual(names, wantNames) {
			t.Errorf(`FullNames(%q) -> %s, want %s`, dir, names, wantNames)
		}
	}
}

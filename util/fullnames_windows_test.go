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
	// Remove filenames that start with ".".
	// XXX: This behavior only serves to make current behavior of FullNames,
	// which always treat dotfiles as hidden, legal; the validness of this
	// behavior is quetionable. However, since FullNames is also depended by the
	// glob package for testing, changing FullNames requires changing the
	// behavior of globbing as well.
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if !strings.HasPrefix(name, dir+".") {
			filtered = append(filtered, name)
		}
	}
	sort.Strings(filtered)
	return filtered
}

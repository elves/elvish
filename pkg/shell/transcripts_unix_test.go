//go:build unix

package shell_test

import "syscall"

func init() {
	sigCHLDName = syscall.SIGCHLD.String()
}

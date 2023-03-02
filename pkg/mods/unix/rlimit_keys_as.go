//go:build linux || freebsd || netbsd

package unix

import "golang.org/x/sys/unix"

func init() {
	addRlimitKeys(map[int]string{
		unix.RLIMIT_AS: "as",
	})
}

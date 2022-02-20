//go:build !openbsd && !darwin && !windows && !plan9 && !js
// +build !openbsd,!darwin,!windows,!plan9,!js

package unix

import "golang.org/x/sys/unix"

func init() {
	addRlimitKeys(map[int]string{
		unix.RLIMIT_AS: "as",
	})
}

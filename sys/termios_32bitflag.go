// +build 386,darwin arm,darwin dragonfly freebsd linux netbsd openbsd solaris

package sys

// The type of Termios.Lflag is different on different platforms.
// This file is for those where Lflag is uint32.

func setFlag(flag *uint32, mask uint32, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

// +build amd64,darwin arm64,darwin

package sys

// The type of Termios.Lflag is different on different platforms.
// This file is for those where Lflag is uint64.

func setFlag(flag *uint64, mask uint64, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

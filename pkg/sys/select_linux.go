// +build linux

package sys

import "golang.org/x/sys/unix"

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet) error {
	// On ARM64, MIPS64 and MIPS64LE, unix.Select is emulated in userland and
	// will dereference timeout. In that case, we use Pselect to work around the
	// problem. Bug: https://github.com/golang/go/issues/24189
	_, err := unix.Pselect(nfd, r.s(), w.s(), e.s(), nil, nil)
	return err
}

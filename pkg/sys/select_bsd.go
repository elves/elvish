// +build darwin dragonfly freebsd netbsd openbsd

package sys

import "golang.org/x/sys/unix"

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet) error {
	return unix.Select(nfd, r.s(), w.s(), e.s(), nil)
}

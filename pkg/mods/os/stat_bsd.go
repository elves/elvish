//go:build darwin || freebsd || netbsd || openbsd

package os

import "syscall"

func init() {
	extraStatFields["gen"] = func(st *syscall.Stat_t) uint64 { return uint64(st.Gen) }
	extraStatFields["flags"] = func(st *syscall.Stat_t) uint64 { return uint64(st.Flags) }
}

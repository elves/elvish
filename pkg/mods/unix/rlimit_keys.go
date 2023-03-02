//go:build unix

package unix

import "golang.org/x/sys/unix"

var rlimitKeys = map[int]string{
	// The following are defined by POSIX
	// (https://pubs.opengroup.org/onlinepubs/9699919799/functions/getrlimit.html).
	//
	// Note: RLIMIT_AS is defined by POSIX, but missing on OpenBSD
	// (https://man.openbsd.org/getrlimit.2#BUGS); it is defined on Darwin, but
	// it's an undocumented alias of RLIMIT_RSS there.
	unix.RLIMIT_CORE:   "core",
	unix.RLIMIT_CPU:    "cpu",
	unix.RLIMIT_DATA:   "data",
	unix.RLIMIT_FSIZE:  "fsize",
	unix.RLIMIT_NOFILE: "nofile",
	unix.RLIMIT_STACK:  "stack",
	// The following are not defined by POSIX, but supported by every Unix OS
	// Elvish supports (Linux, macOS, Free/Net/OpenBSD). See:
	//
	// - https://man7.org/linux/man-pages/man2/setrlimit.2.html
	// - https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/getrlimit.2.html
	// - https://www.freebsd.org/cgi/man.cgi?query=getrlimit
	// - https://man.netbsd.org/getrlimit.2
	// - https://man.openbsd.org/getrlimit.2
	unix.RLIMIT_MEMLOCK: "memlock",
	unix.RLIMIT_NPROC:   "nproc",
	unix.RLIMIT_RSS:     "rss",
}

//lint:ignore U1000 used on some OS
func addRlimitKeys(m map[int]string) {
	for k, v := range m {
		rlimitKeys[k] = v
	}
}

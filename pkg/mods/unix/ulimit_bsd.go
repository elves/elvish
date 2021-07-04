//go:build darwin || dragonfly || netbsd || openbsd
// +build darwin dragonfly netbsd openbsd

package unix

import "golang.org/x/sys/unix"

var ulimitResources = map[string]ulimitResource{
	"as":      {"virtual memory", "bytes", unix.RLIMIT_AS},
	"core":    {"core file size", "bytes", unix.RLIMIT_CORE},
	"cpu":     {"cpu time", "seconds", unix.RLIMIT_CPU},
	"data":    {"data segment size", "bytes", unix.RLIMIT_DATA},
	"fsize":   {"file size", "bytes", unix.RLIMIT_FSIZE},
	"memlock": {"max locked memory", "bytes", unix.RLIMIT_MEMLOCK},
	"nofile":  {"open files", "count", unix.RLIMIT_NOFILE},
	"nproc":   {"max user processes", "count", unix.RLIMIT_NPROC},
	"rss":     {"max memory size", "bytes", unix.RLIMIT_RSS},
	"stack":   {"stack size", "bytes", unix.RLIMIT_STACK},
}

func SetRlimitCur(rlimit *unix.Rlimit, newLimit uint64) error {
	rlimit.Cur = newLimit
	return nil
}

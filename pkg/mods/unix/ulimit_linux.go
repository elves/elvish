//go:build linux
// +build linux

package unix

import "golang.org/x/sys/unix"

var ulimitResources = map[string]ulimitResource{
	"as":         {"virtual memory", "bytes", unix.RLIMIT_AS},
	"core":       {"core file size", "bytes", unix.RLIMIT_CORE},
	"cpu":        {"cpu time", "seconds", unix.RLIMIT_CPU},
	"data":       {"data segment size", "bytes", unix.RLIMIT_DATA},
	"fsize":      {"file size", "bytes", unix.RLIMIT_FSIZE},
	"locks":      {"file locks", "count", unix.RLIMIT_LOCKS},
	"memlock":    {"max locked memory", "bytes", unix.RLIMIT_MEMLOCK},
	"msgqueue":   {"POSIX message queues", "count", unix.RLIMIT_MSGQUEUE},
	"nice":       {"nice limit", "count", unix.RLIMIT_NICE},
	"nofile":     {"open files", "count", unix.RLIMIT_NOFILE},
	"nproc":      {"max user processes", "count", unix.RLIMIT_NPROC},
	"rss":        {"max memory size", "bytes", unix.RLIMIT_RSS},
	"rtprio":     {"real-time priority", "count", unix.RLIMIT_RTPRIO},
	"rttime":     {"real-time limit", "microseconds", unix.RLIMIT_RTTIME},
	"sigpending": {"pending signals", "count", unix.RLIMIT_SIGPENDING},
	"stack":      {"stack size", "bytes", unix.RLIMIT_STACK},
}

func SetRlimitCur(rlimit *unix.Rlimit, newLimit uint64) error {
	rlimit.Cur = newLimit
	return nil
}

package unix

import "golang.org/x/sys/unix"

func init() {
	// https://man7.org/linux/man-pages/man2/setrlimit.2.html
	addRlimitKeys(map[int]string{
		unix.RLIMIT_LOCKS:      "locks",
		unix.RLIMIT_MSGQUEUE:   "msgqueue",
		unix.RLIMIT_NICE:       "nice",
		unix.RLIMIT_RTPRIO:     "rtprio",
		unix.RLIMIT_RTTIME:     "rttime",
		unix.RLIMIT_SIGPENDING: "sigpending",
	})
}

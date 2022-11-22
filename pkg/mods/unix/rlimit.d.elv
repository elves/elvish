# A map describing resource limits of the current process.
#
# Each key is a string corresponds to a resource, and each value is a map with
# keys `&cur` and `&max`, describing the soft and hard limits of that resource.
# A missing `&cur` key means that there is no soft limit; a missing `&max` key
# means that there is no hard limit.
#
# The following resources are supported, some only present on certain OSes:
#
# | Key          | Resource           | Unit    | OS                 |
# | ------------ | ------------------ | ------- | ------------------ |
# | `core`       | Core file          | bytes   | all                |
# | `cpu`        | CPU time           | seconds | all                |
# | `data`       | Data segment       | bytes   | all                |
# | `fsize`      | File size          | bytes   | all                |
# | `memlock`    | Locked memory      | bytes   | all                |
# | `nofile`     | File descriptors   | number  | all                |
# | `nproc`      | Processes          | number  | all                |
# | `rss`        | Resident set size  | bytes   | all                |
# | `stack`      | Stack segment      | bytes   | all                |
# | `as`         | Address space      | bytes   | Linux, Free/NetBSD |
# | `nthr`       | Threads            | number  | NetBSD             |
# | `sbsize`     | Socket buffers     | bytes   | NetBSD             |
# | `locks`      | File locks         | number  | Linux              |
# | `msgqueue`   | Message queues     | bytes   | Linux              |
# | `nice`       | 20 - nice value    |         | Linux              |
# | `rtprio`     | Real-time priority |         | Linux              |
# | `rttime`     | Real-time CPU time | seconds | Linux              |
# | `sigpending` | Signals queued     | number  | Linux              |
#
# For the exact semantics of each resource, see the man page of `getrlimit`:
# [Linux](https://man7.org/linux/man-pages/man2/setrlimit.2.html),
# [macOS](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/getrlimit.2.html),
# [FreeBSD](https://www.freebsd.org/cgi/man.cgi?query=getrlimit),
# [NetBSD](https://man.netbsd.org/getrlimit.2),
# [OpenBSD](https://man.openbsd.org/getrlimit.2). A key `foo` in the Elvish API
# corresponds to `RLIMIT_FOO` in the C API.
#
# Examples:
#
# ```elvish-transcript
# ~> put $unix:rlimits
# ▶ [&nofile=[&cur=(num 256)] &fsize=[&] &nproc=[&max=(num 2666) &cur=(num 2666)] &memlock=[&] &cpu=[&] &core=[&cur=(num 0)] &stack=[&max=(num 67092480) &cur=(num 8372224)] &rss=[&] &data=[&]]
# ~> # mimic Bash's "ulimit -a"
# ~> keys $unix:rlimits | order | each {|key|
#      var m = $unix:rlimits[$key]
#      fn get {|k| if (has-key $m $k) { put $m[$k] } else { put unlimited } }
#      printf "%-7v %-9v %-9v\n" $key (get cur) (get max)
#    }
# core    0         unlimited
# cpu     unlimited unlimited
# data    unlimited unlimited
# fsize   unlimited unlimited
# memlock unlimited unlimited
# nofile  256       unlimited
# nproc   2666      2666
# rss     unlimited unlimited
# stack   8372224   67092480
# ~> # Decrease the soft limit on file descriptors
# ~> set unix:rlimits[nofile][cur] = 100
# ~> put $unix:rlimits[nofile]
# ▶ [&cur=(num 100)]
# ~> # Remove the soft limit on file descriptors
# ~> del unix:rlimits[nofile][cur]
# ~> put $unix:rlimits[nofile]
# ▶ [&]
# ```
var rlimits

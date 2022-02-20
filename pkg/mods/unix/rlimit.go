//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package unix

import (
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

//elvdoc:var rlimits
//
// A map describing resource limits of the current process.
//
// Each key is a string corresponds to a resource, and each value is a map with
// keys `&cur` and `&max`, describing the soft and hard limits of that resource.
// A missing `&cur` key means that there is no soft limit; a missing `&max` key
// means that there is no hard limit.
//
// The following resources are supported, some only present on certain OSes:
//
// | Key          | Resource           | Unit    | OS                 |
// | ------------ | ------------------ | ------- | ------------------ |
// | `core`       | Core file          | bytes   | all                |
// | `cpu`        | CPU time           | seconds | all                |
// | `data`       | Data segment       | bytes   | all                |
// | `fsize`      | File size          | bytes   | all                |
// | `memlock`    | Locked memory      | bytes   | all                |
// | `nofile`     | File descriptors   | number  | all                |
// | `nproc`      | Processes          | number  | all                |
// | `rss`        | Resident set size  | bytes   | all                |
// | `stack`      | Stack segment      | bytes   | all                |
// | `as`         | Address space      | bytes   | Linux, Free/NetBSD |
// | `nthr`       | Threads            | number  | NetBSD             |
// | `sbsize`     | Socket buffers     | bytes   | NetBSD             |
// | `locks`      | File locks         | number  | Linux              |
// | `msgqueue`   | Message queues     | bytes   | Linux              |
// | `nice`       | 20 - nice value    |         | Linux              |
// | `rtprio`     | Real-time priority |         | Linux              |
// | `rttime`     | Real-time CPU time | seconds | Linux              |
// | `sigpending` | Signals queued     | number  | Linux              |
//
// For the exact semantics of each resource, see the man page of `getrlimit`:
// [Linux](https://man7.org/linux/man-pages/man2/setrlimit.2.html),
// [macOS](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/getrlimit.2.html),
// [FreeBSD](https://www.freebsd.org/cgi/man.cgi?query=getrlimit),
// [NetBSD](https://man.netbsd.org/getrlimit.2),
// [OpenBSD](https://man.openbsd.org/getrlimit.2). A key `foo` in the Elvish API
// corresponds to `RLIMIT_FOO` in the C API.
//
// Examples:
//
// ```elvish-transcript
// ~> put $unix:rlimits
// ▶ [&nofile=[&cur=(num 256)] &fsize=[&] &nproc=[&max=(num 2666) &cur=(num 2666)] &memlock=[&] &cpu=[&] &core=[&cur=(num 0)] &stack=[&max=(num 67092480) &cur=(num 8372224)] &rss=[&] &data=[&]]
// ~> # mimic Bash's "ulimit -a"
// ~> keys $unix:rlimits | order | each {|key|
//      var m = $unix:rlimits[$key]
//      fn get {|k| if (has-key $m $k) { put $m[$k] } else { put unlimited } }
//      printf "%-7v %-9v %-9v\n" $key (get cur) (get max)
//    }
// core    0         unlimited
// cpu     unlimited unlimited
// data    unlimited unlimited
// fsize   unlimited unlimited
// memlock unlimited unlimited
// nofile  256       unlimited
// nproc   2666      2666
// rss     unlimited unlimited
// stack   8372224   67092480
// ~> # Decrease the soft limit on file descriptors
// ~> set unix:rlimits[nofile][cur] = 100
// ~> put $unix:rlimits[nofile]
// ▶ [&cur=(num 100)]
// ~> # Remove the soft limit on file descriptors
// ~> del unix:rlimits[nofile][cur]
// ~> put $unix:rlimits[nofile]
// ▶ [&]
// ```

type rlimitsVar struct{}

var (
	getRlimit = unix.Getrlimit
	setRlimit = unix.Setrlimit
)

var (
	rlimitMutex sync.Mutex
	rlimits     map[int]*unix.Rlimit
)

func (rlimitsVar) Get() interface{} {
	rlimitMutex.Lock()
	defer rlimitMutex.Unlock()

	initRlimits()
	rlimitsMap := vals.EmptyMap
	for res, lim := range rlimits {
		limMap := vals.EmptyMap
		if lim.Cur != unix.RLIM_INFINITY {
			limMap = limMap.Assoc("cur", convertRlimT(lim.Cur))
		}
		if lim.Max != unix.RLIM_INFINITY {
			limMap = limMap.Assoc("max", convertRlimT(lim.Max))
		}
		rlimitsMap = rlimitsMap.Assoc(rlimitKeys[res], limMap)
	}
	return rlimitsMap
}

func (rlimitsVar) Set(v interface{}) error {
	newRlimits, err := parseRlimitsMap(v)
	if err != nil {
		return err
	}

	rlimitMutex.Lock()
	defer rlimitMutex.Unlock()

	initRlimits()
	for res := range rlimits {
		if *rlimits[res] != *newRlimits[res] {
			err := setRlimit(res, newRlimits[res])
			if err != nil {
				return fmt.Errorf("setrlimit %s: %w", rlimitKeys[res], err)
			}
			rlimits[res] = newRlimits[res]
		}
	}
	return nil
}

func initRlimits() {
	if rlimits != nil {
		return
	}
	rlimits = make(map[int]*unix.Rlimit)
	for res := range rlimitKeys {
		var lim unix.Rlimit
		err := getRlimit(res, &lim)
		if err == nil {
			rlimits[res] = &lim
		} else {
			// Since getrlimit should only ever return an error when the
			// resource is not supported, this should normally never happen. But
			// be defensive nonetheless.
			logger.Println("initialize rlimits", res, rlimitKeys[res], err)
			// Remove this key, so that rlimitKeys is always consistent with the
			// value of rlimits (and thus $unix:rlimits).
			delete(rlimitKeys, res)
		}
	}
}

func parseRlimitsMap(val interface{}) (map[int]*unix.Rlimit, error) {
	if err := checkRlimitsMapKeys(val); err != nil {
		return nil, err
	}
	limits := make(map[int]*unix.Rlimit, len(rlimitKeys))
	for res, key := range rlimitKeys {
		limitVal, err := vals.Index(val, key)
		if err != nil {
			return nil, err
		}
		limits[res], err = parseRlimitMap(limitVal)
		if err != nil {
			return nil, err
		}
	}
	return limits, nil
}

func checkRlimitsMapKeys(val interface{}) error {
	wantedKeys := make(map[string]struct{}, len(rlimitKeys))
	for _, key := range rlimitKeys {
		wantedKeys[key] = struct{}{}
	}
	var errKey error
	err := vals.IterateKeys(val, func(k interface{}) bool {
		ks, ok := k.(string)
		if !ok {
			errKey = errs.BadValue{What: "key of $unix:rlimits",
				Valid: "string", Actual: vals.Kind(k)}
			return false
		}
		if _, valid := wantedKeys[ks]; !valid {
			errKey = errs.BadValue{What: "key of $unix:rlimits",
				Valid: "valid resource key", Actual: vals.ReprPlain(k)}
			return false
		}
		delete(wantedKeys, ks)
		return true
	})
	if err != nil {
		return errs.BadValue{What: "$unix:rlimits",
			Valid: "map", Actual: vals.Kind(val)}
	}
	if errKey != nil {
		return errKey
	}
	if len(wantedKeys) > 0 {
		return errs.BadValue{What: "$unix:rlimits",
			Valid: "map containing all resource keys", Actual: vals.ReprPlain(val)}
	}
	return nil
}

func parseRlimitMap(val interface{}) (*unix.Rlimit, error) {
	if err := checkRlimitMapKeys(val); err != nil {
		return nil, err
	}
	cur, err := indexRlimitMap(val, "cur")
	if err != nil {
		return nil, err
	}
	max, err := indexRlimitMap(val, "max")
	if err != nil {
		return nil, err
	}
	return &unix.Rlimit{Cur: cur, Max: max}, nil
}

func checkRlimitMapKeys(val interface{}) error {
	var errKey error
	err := vals.IterateKeys(val, func(k interface{}) bool {
		if k != "cur" && k != "max" {
			errKey = errs.BadValue{What: "key of rlimit value",
				Valid: "cur or max", Actual: vals.ReprPlain(k)}
			return false
		}
		return true
	})
	if err != nil {
		return errs.BadValue{What: "rlimit value",
			Valid: "map", Actual: vals.Kind(val)}
	}
	return errKey
}

func indexRlimitMap(m interface{}, key string) (rlimT, error) {
	val, err := vals.Index(m, key)
	if err != nil {
		return unix.RLIM_INFINITY, nil
	}
	if r, ok := parseRlimT(val); ok {
		return r, nil
	}
	return 0, errs.BadValue{What: key + " in rlimit value",
		Valid: rlimTValid, Actual: vals.ReprPlain(val)}
}

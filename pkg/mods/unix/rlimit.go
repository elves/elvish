//go:build unix

package unix

import (
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

type rlimitsVar struct{}

var (
	getRlimit = unix.Getrlimit
	setRlimit = unix.Setrlimit
)

var (
	rlimitMutex sync.Mutex
	rlimits     map[int]*unix.Rlimit
)

func (rlimitsVar) Get() any {
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

func (rlimitsVar) Set(v any) error {
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

func parseRlimitsMap(val any) (map[int]*unix.Rlimit, error) {
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

func checkRlimitsMapKeys(val any) error {
	wantedKeys := make(map[string]struct{}, len(rlimitKeys))
	for _, key := range rlimitKeys {
		wantedKeys[key] = struct{}{}
	}
	var errKey error
	err := vals.IterateKeys(val, func(k any) bool {
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

func parseRlimitMap(val any) (*unix.Rlimit, error) {
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

func checkRlimitMapKeys(val any) error {
	var errKey error
	err := vals.IterateKeys(val, func(k any) bool {
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

func indexRlimitMap(m any, key string) (rlimT, error) {
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

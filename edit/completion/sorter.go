package completion

import (
	"errors"
	"os"
	"sort"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/xiaq/persistent/hashmap"
)

var (
	errSorterMustBeFn = errors.New("sorter must be a function")
)

type rawCandidates []rawCandidate

func (cs rawCandidates) Len() int           { return len(cs) }
func (cs rawCandidates) Swap(i, j int)      { cs[i], cs[j] = cs[j], cs[i] }
func (cs rawCandidates) Less(i, j int) bool { return cs[i].text() < cs[j].text() }

func lookupSorter(m hashmap.Map, name string) (eval.Callable, bool) {
	key := name
	if !hashmap.HasKey(m, key) {
		// Use fallback sorter
		if !hashmap.HasKey(m, "") {
			return nil, false
		}
		key = ""
	}
	value, _ := m.Index(key)
	sorter, ok := value.(eval.Callable)
	return sorter, ok
}

func sortRawCandidates(ev *eval.Evaler, customSorter eval.Callable, seed string, rs []rawCandidate) ([]rawCandidate, error) {
	// default sorter
	if customSorter == nil {
		sort.Sort(rawCandidates(rs))
		return rs, nil
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}
	ec := eval.NewTopFrame(ev, eval.NewInternalSource("[editor sorter]"), ports)
	sort.SliceStable(rs, func(i, j int) bool {
		args := []interface{}{seed, rs[i].text(), rs[j].text()}
		rets, err := ec.CaptureOutput(customSorter, args, eval.NoOpts)
		if err != nil {
			logger.Printf("sorter: %v", err)
			return false
		}
		// just return one bool
		if n := len(rets); n != 1 {
			logger.Printf("sorter: return value count %d != 1", n)
			return false
		}
		return vals.Bool(rets[0])
	})

	return rs, nil
}

package history

import (
	"sync"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

// List is a list-like value that provides access to history in the API
// backend. It is used in the $edit:history variable.
type List struct {
	*sync.RWMutex
	Daemon *daemon.Client
}

func (hv List) Kind() string {
	return "list"
}

// Equal returns true as long as the rhs is also of type History.
func (hv List) Equal(a interface{}) bool {
	_, ok := a.(List)
	return ok
}

func (hv List) Hash() uint32 {
	// TODO(xiaq): Make a global registry of singleton hashes to avoid
	// collision.
	return 100
}

func (hv List) Repr(int) string {
	return "$edit:history"
}

func (hv List) Len() int {
	hv.RLock()
	defer hv.RUnlock()

	nextseq, err := hv.Daemon.NextCmdSeq()
	maybeThrow(err)
	return nextseq - 1
}

func (hv List) Iterate(f func(interface{}) bool) {
	hv.RLock()
	defer hv.RUnlock()

	n := hv.Len()
	cmds, err := hv.Daemon.Cmds(1, n+1)
	maybeThrow(err)

	for _, cmd := range cmds {
		if !f(string(cmd)) {
			break
		}
	}
}

func (hv List) Index(rawIndex interface{}) (interface{}, error) {
	hv.RLock()
	defer hv.RUnlock()

	index, err := vals.ConvertListIndex(rawIndex, hv.Len())
	if err != nil {
		return nil, err
	}
	if index.Slice {
		cmds, err := hv.Daemon.Cmds(index.Lower+1, index.Upper+1)
		if err != nil {
			return nil, err
		}
		l := vals.EmptyList
		for _, cmd := range cmds {
			l = l.Cons(cmd)
		}
		return l, nil
	}
	s, err := hv.Daemon.Cmd(index.Lower + 1)
	return string(s), err
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

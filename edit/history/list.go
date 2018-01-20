package history

import (
	"sync"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
)

// List is a list-like value that provides access to history in the API
// backend. It is used in the $edit:history variable.
type List struct {
	*sync.RWMutex
	Daemon *daemon.Client
}

var (
	_ types.Value    = List{}
	_ types.ListLike = List{}
)

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

func (hv List) Iterate(f func(types.Value) bool) {
	hv.RLock()
	defer hv.RUnlock()

	n := hv.Len()
	cmds, err := hv.Daemon.Cmds(1, n+1)
	maybeThrow(err)

	for _, cmd := range cmds {
		if !f(types.String(cmd)) {
			break
		}
	}
}

func (hv List) Index(idx types.Value) (types.Value, error) {
	hv.RLock()
	defer hv.RUnlock()

	slice, i, j, err := types.ParseAndFixListIndex(types.ToString(idx), hv.Len())
	if err != nil {
		return nil, err
	}
	if slice {
		cmds, err := hv.Daemon.Cmds(i+1, j+1)
		if err != nil {
			return nil, err
		}
		vs := make([]types.Value, len(cmds))
		for i := range cmds {
			vs[i] = types.String(cmds[i])
		}
		return types.MakeList(vs...), nil
	}
	s, err := hv.Daemon.Cmd(i + 1)
	return types.String(s), err
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

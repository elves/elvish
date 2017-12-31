package history

import (
	"sync"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
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
	_ types.Value   = List{}
	_ eval.ListLike = List{}
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
		if !f(eval.String(cmd)) {
			break
		}
	}
}

func (hv List) IndexOne(idx types.Value) types.Value {
	hv.RLock()
	defer hv.RUnlock()

	slice, i, j := eval.ParseAndFixListIndex(eval.ToString(idx), hv.Len())
	if slice {
		cmds, err := hv.Daemon.Cmds(i+1, j+1)
		maybeThrow(err)
		vs := make([]types.Value, len(cmds))
		for i := range cmds {
			vs[i] = eval.String(cmds[i])
		}
		return eval.NewList(vs...)
	}
	s, err := hv.Daemon.Cmd(i + 1)
	maybeThrow(err)
	return eval.String(s)
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

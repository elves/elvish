package edit

import (
	"sync"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
)

// History implements the $le:history variable. It is list-like.
type History struct {
	mutex *sync.RWMutex
	st    *store.Store
}

var _ eval.ListLike = History{}

func (hv History) Kind() string {
	return "list"
}

func (hv History) Repr(int) string {
	return "$le:history"
}

func (hv History) Len() int {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	nextseq, err := hv.st.NextCmdSeq()
	maybeThrow(err)
	return nextseq - 1
}

func (hv History) Iterate(f func(eval.Value) bool) {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	n := hv.Len()
	err := hv.st.IterateCmds(1, n+1, func(cmd string) bool {
		return f(eval.String(cmd))
	})
	maybeThrow(err)
}

func (hv History) IndexOne(idx eval.Value) eval.Value {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	slice, i, j := eval.ParseAndFixListIndex(eval.ToString(idx), hv.Len())
	if slice {
		cmds, err := hv.st.Cmds(i+1, j+1)
		maybeThrow(err)
		vs := make([]eval.Value, len(cmds))
		for i := range cmds {
			vs[i] = eval.String(cmds[i])
		}
		return eval.NewList(vs...)
	}
	s, err := hv.st.Cmd(i + 1)
	maybeThrow(err)
	return eval.String(s)
}

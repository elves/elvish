package edit

import (
	"sync"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
)

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
	for i := 1; i <= n; i++ {
		s, err := hv.st.Cmd(i)
		maybeThrow(err)
		if !f(eval.String(s)) {
			return
		}
	}
}

func (hv History) IndexOne(idx eval.Value) eval.Value {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	slice, i, j := eval.ParseAndFixListIndex(eval.ToString(idx), hv.Len())
	if slice {
		ss := make([]eval.Value, j-i)
		for k := i + 1; k < j+1; k++ {
			s, err := hv.st.Cmd(k)
			maybeThrow(err)
			ss[k-(i+1)] = eval.String(s)
		}
		return eval.NewList(ss...)
	}
	s, err := hv.st.Cmd(i + 1)
	maybeThrow(err)
	return eval.String(s)
}

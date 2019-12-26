package store

import (
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/store"
)

func Ns(s store.Store) eval.Ns {
	return eval.NewNs().AddGoFns("store:", map[string]interface{}{
		"del-dir": s.DelDir,
		"del-cmd": s.DelCmd,
	})
}

package store

import (
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/store/storedefs"
)

func Ns(s storedefs.Store) eval.Ns {
	return eval.NewNs().AddGoFns("store:", map[string]interface{}{
		"del-dir": s.DelDir,
		"del-cmd": s.DelCmd,
	})
}

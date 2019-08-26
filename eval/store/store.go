package store

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

func Ns(s storedefs.Store) eval.Ns {
	return eval.NewNs().AddGoFns("store:", map[string]interface{}{
		"cmds":     s.Cmds,
		"next-cmd": s.NextCmd,
		"del-cmd":  s.DelCmd,
		"del-dir":  s.DelDir,
	})
}

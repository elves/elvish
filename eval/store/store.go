package store

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

func Ns(s storedefs.Store) eval.Ns {
	return eval.NewNs().AddBuiltinFns("store:", map[string]interface{}{
		"del-dir": s.DelDir,
		"del-cmd": s.DelCmd,
	})
}

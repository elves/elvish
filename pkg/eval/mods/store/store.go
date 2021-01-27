package store

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/store"
)

func Ns(s store.Store) *eval.Ns {
	return eval.NsBuilder{}.AddGoFns("store:", map[string]interface{}{
		"del-dir": s.DelDir,
		"del-cmd": s.DelCmd,
	}).Ns()
}

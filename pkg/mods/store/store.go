package store

import (
	_ "embed"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/store/storedefs"
)

func Ns(s storedefs.Store) *eval.Ns {
	return eval.BuildNsNamed("store").
		AddGoFns(map[string]any{
			"next-cmd-seq": s.NextCmdSeq,
			"add-cmd":      s.AddCmd,
			"del-cmd":      s.DelCmd,
			"cmd":          s.Cmd,
			"cmds":         s.CmdsWithSeq,
			"next-cmd":     s.NextCmd,
			"prev-cmd":     s.PrevCmd,

			"add-dir": func(dir string) error { return s.AddDir(dir, 1) },
			"del-dir": s.DelDir,
			"dirs":    func() ([]storedefs.Dir, error) { return s.Dirs(storedefs.NoBlacklist) },
		}).Ns()
}

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

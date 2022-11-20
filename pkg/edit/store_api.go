package edit

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse/parseutil"
	"src.elv.sh/pkg/store/storedefs"
)

var errStoreOffline = errors.New("store offline")

type cmdhistOpt struct{ CmdOnly, Dedup, NewestFirst bool }

func (o *cmdhistOpt) SetDefaultOptions() {}

func commandHistory(opts cmdhistOpt, fuser histutil.Store, out eval.ValueOutput) error {
	if fuser == nil {
		return errStoreOffline
	}
	cmds, err := fuser.AllCmds()
	if err != nil {
		return err
	}
	if opts.Dedup {
		cmds = dedupCmds(cmds, opts.NewestFirst)
	} else if opts.NewestFirst {
		reverseCmds(cmds)
	}
	if opts.CmdOnly {
		for _, cmd := range cmds {
			err := out.Put(cmd.Text)
			if err != nil {
				return err
			}
		}
	} else {
		for _, cmd := range cmds {
			err := out.Put(vals.MakeMap("id", cmd.Seq, "cmd", cmd.Text))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dedupCmds(allCmds []storedefs.Cmd, newestFirst bool) []storedefs.Cmd {
	// Capacity allocation below is based on some personal empirical observation.
	uniqCmds := make([]storedefs.Cmd, 0, len(allCmds)/4)
	seenCmds := make(map[string]bool, len(allCmds)/4)
	for i := len(allCmds) - 1; i >= 0; i-- {
		if !seenCmds[allCmds[i].Text] {
			seenCmds[allCmds[i].Text] = true
			uniqCmds = append(uniqCmds, allCmds[i])
		}
	}
	if !newestFirst {
		reverseCmds(uniqCmds)
	}
	return uniqCmds
}

// Reverse the order of commands, in place, in the slice. This reorders the
// command history between oldest or newest command being first in the slice.
func reverseCmds(cmds []storedefs.Cmd) {
	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
}

func insertLastWord(app cli.App, histStore histutil.Store) error {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return nil
	}
	c := histStore.Cursor("")
	c.Prev()
	cmd, err := c.Get()
	if err != nil {
		return err
	}
	words := parseutil.Wordify(cmd.Text)
	if len(words) > 0 {
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.Buffer.InsertAtDot(words[len(words)-1])
		})
	}
	return nil
}

func initStoreAPI(app cli.App, nb eval.NsBuilder, fuser histutil.Store) {
	nb.AddGoFns(map[string]any{
		"command-history": func(fm *eval.Frame, opts cmdhistOpt) error {
			return commandHistory(opts, fuser, fm.ValueOutput())
		},
		"insert-last-word": func() { insertLastWord(app, fuser) },
	})
}

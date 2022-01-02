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

//elvdoc:fn command-history
//
// ```elvish
// edit:command-history &cmd-only=$false &dedup=$false &newest-first
// ```
//
// Outputs the command history.
//
// By default, each entry is represented as a map, with an `id` key key for the
// sequence number of the command, and a `cmd` key for the text of the command.
// If `&cmd-only` is `$true`, only the text of each command is output.
//
// All entries are output by default. If `&dedup` is `$true`, only the most
// recent instance of each command (when comparing just the `cmd` key) is
// output.
//
// Commands are are output in oldest to newest order by default. If
// `&newest-first` is `$true` the output is in newest to oldest order instead.
//
// As an example, either of the following extracts the text of the most recent
// command:
//
// ```elvish
// edit:command-history | put [(all)][-1][cmd]
// edit:command-history &cmd-only &newest-first | take 1
// ```

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

//elvdoc:fn insert-last-word
//
// Inserts the last word of the last command.

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
	nb.AddGoFns(map[string]interface{}{
		"command-history": func(fm *eval.Frame, opts cmdhistOpt) error {
			return commandHistory(opts, fuser, fm.ValueOutput())
		},
		"insert-last-word": func() { insertLastWord(app, fuser) },
	})
}

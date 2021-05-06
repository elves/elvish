package edit

import (
	"errors"
	"strconv"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse/parseutil"
	"src.elv.sh/pkg/store"
)

var errStoreOffline = errors.New("store offline")

//elvdoc:fn command-history
//
// ```elvish
// edit:command-history &dedup &newest-first
// ```
//
// Outputs the command history as a stream of maps. Each map has a `id` key for the sequence number
// of the command, and a `cmd` key for the text of the command. If `&dedup` is true only the most
// recent instance of each command (when comparing just the `cmd` key) is output. If `&newest-first`
// is true the output is in newest to oldest order (default is oldest to newest).
//
// Use indexing to extract individual entries. For example, to extract the text of the most recent
// command do this:
//
// ```elvish
// edit:command-history | put [(all)][-1][cmd]
// ```
//
// @cf builtin:dir-history

type cmdhistOpt struct{ Dedup, NewestFirst bool }

func (o *cmdhistOpt) SetDefaultOptions() {}

func commandHistory(opts cmdhistOpt, fuser histutil.Store, ch chan<- interface{}) error {
	if fuser == nil {
		return errStoreOffline
	}
	cmds, err := fuser.AllCmds()
	if err != nil {
		return err
	}
	if opts.Dedup {
		cmds = uniqCmds(cmds, opts.NewestFirst)
	} else if opts.NewestFirst {
		reverse(cmds)
	}
	for _, cmd := range cmds {
		ch <- vals.MakeMap("id", strconv.Itoa(cmd.Seq), "cmd", cmd.Text)
	}
	return nil
}

func uniqCmds(allCmds []store.Cmd, newestFirst bool) []store.Cmd {
	// The division by four is based on the empirical observation that my personal command history
	// tends to be only 25% unique commands. This is merely a hint to minimize allocations.
	uniqCmds := make([]store.Cmd, 0, len(allCmds)/4)
	seenCmds := make(map[string]bool, len(allCmds)/4)
	for i := len(allCmds) - 1; i != 0; i-- {
		if !seenCmds[allCmds[i].Text] {
			seenCmds[allCmds[i].Text] = true
			uniqCmds = append(uniqCmds, allCmds[i])
		}
	}
	if !newestFirst {
		reverse(uniqCmds)
	}
	return uniqCmds
}

// Reverse the order of commands, in place, in the slice. This reorders the command history between
// oldest or newest command being first in the slice.
func reverse(cmds []store.Cmd) {
	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
}

//elvdoc:fn insert-last-word
//
// Inserts the last word of the last command.

func insertLastWord(app cli.App, histStore histutil.Store) error {
	c := histStore.Cursor("")
	c.Prev()
	cmd, err := c.Get()
	if err != nil {
		return err
	}
	words := parseutil.Wordify(cmd.Text)
	if len(words) > 0 {
		app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
			s.Buffer.InsertAtDot(words[len(words)-1])
		})
	}
	return nil
}

func initStoreAPI(app cli.App, nb eval.NsBuilder, fuser histutil.Store) {
	nb.AddGoFns("<edit>", map[string]interface{}{
		"command-history": func(fm *eval.Frame, opts cmdhistOpt) error {
			return commandHistory(opts, fuser, fm.OutputChan())
		},
		"insert-last-word": func() { insertLastWord(app, fuser) },
	})
}

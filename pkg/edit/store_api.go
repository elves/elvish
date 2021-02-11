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
)

var errStoreOffline = errors.New("store offline")

//elvdoc:fn command-history
//
// Outputs the entire command history as a stream of maps. Each map has a `id`
// key that identifies the sequence number of the entry, and a `cmd` key that
// identifies the content.
//
// Use indexing to extract individual entries. For example, to extract the
// content of the last command, do this:
//
// ```elvish
// edit:command-history | put [(all)][-1][cmd]
// ```

func commandHistory(fuser histutil.Store, ch chan<- interface{}) error {
	if fuser == nil {
		return errStoreOffline
	}
	cmds, err := fuser.AllCmds()
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		ch <- vals.MakeMap("id", strconv.Itoa(cmd.Seq), "cmd", cmd.Text)
	}
	return nil
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
		"command-history": func(fm *eval.Frame) error {
			return commandHistory(fuser, fm.OutputChan())
		},
		"insert-last-word": func() { insertLastWord(app, fuser) },
	})
}

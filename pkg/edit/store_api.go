package edit

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse/parseutil"
)

var errStoreOffline = errors.New("store offline")

//elvdoc:fn command-history
//
// Outputs the entire command history.

func commandHistory(fuser *histutil.Fuser, ch chan<- interface{}) error {
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

func insertLastWord(app cli.App, fuser *histutil.Fuser) error {
	if fuser == nil {
		return errStoreOffline
	}
	cmd, err := fuser.LastCmd()
	if err != nil {
		return err
	}
	words := parseutil.Wordify(cmd.Text)
	if len(words) > 0 {
		app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
			s.Buffer.InsertAtDot(words[len(words)-1])
		})
	}
	return nil
}

func initStoreAPI(app cli.App, ns eval.Ns, fuser *histutil.Fuser) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"command-history": func(fm *eval.Frame) error {
			return commandHistory(fuser, fm.OutputChan())
		},
		"insert-last-word": func() { insertLastWord(app, fuser) },
	})
}

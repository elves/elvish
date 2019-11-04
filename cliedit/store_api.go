package cliedit

import (
	"errors"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/store/storedefs"
)

var errStoreOffline = errors.New("store offline")

//elvdoc:fn insert-last-word
//
// Inserts the last word of the last command.

func insertLastWord(app cli.App, st storedefs.Store) error {
	if st == nil {
		return errStoreOffline
	}
	_, cmd, err := st.PrevCmd(-1, "")
	if err != nil {
		return err
	}
	words := parseutil.Wordify(cmd)
	if len(words) > 0 {
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.Buffer.InsertAtDot(words[len(words)-1])
		})
	}
	return nil
}

func initStoreAPI(app cli.App, ns eval.Ns, st storedefs.Store) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"insert-last-word": func() { insertLastWord(app, st) },
	})
}

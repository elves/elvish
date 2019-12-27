package lastcmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/cli"
	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/ui"
)

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) LastCmd() (store.Cmd, error) {
	return store.Cmd{}, mockError
}

func TestStart_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{})
	f.TestTTYNotes(t, "no history store")
}

func TestStart_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{Store: faultyStore{}})
	f.TestTTYNotes(t, "db error: mock error")
}

func TestStart_OK(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemoryStore()
	st.AddCmd(store.Cmd{Text: "foo,bar,baz", Seq: 0})
	Start(f.App, Config{
		Store: st,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})

	// Test UI.
	f.TestTTY(t,
		"\n", // empty code area
		" LASTCMD  ", Styles,
		"********* ", term.DotHere, "\n",
		"    foo,bar,baz                                   \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"  0 foo\n",
		"  1 bar\n",
		"  2 baz",
	)

	// Test negative filtering.
	f.TTY.Inject(term.K('-'))
	f.TestTTY(t,
		"\n", // empty code area
		" LASTCMD  -", Styles,
		"*********  ", term.DotHere, "\n",
		" -3 foo                                           \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		" -2 bar\n",
		" -1 baz",
	)

	// Test automatic submission.
	f.TTY.Inject(term.K('2')) // -2 bar
	f.TestTTY(t, "bar", term.DotHere)

	// Test submission by Enter.
	f.App.CodeArea().MutateState(func(s *cli.CodeAreaState) {
		*s = cli.CodeAreaState{}
	})
	Start(f.App, Config{
		Store: st,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo,bar,baz", term.DotHere)

	// Default wordifier.
	f.App.CodeArea().MutateState(func(s *cli.CodeAreaState) {
		*s = cli.CodeAreaState{}
	})
	st.AddCmd(store.Cmd{Text: "foo bar baz", Seq: 1})
	Start(f.App, Config{Store: st})
	f.TTY.Inject(term.K('0'))
	f.TestTTY(t, "foo", term.DotHere)
}

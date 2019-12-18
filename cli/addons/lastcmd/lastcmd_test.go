package lastcmd

import (
	"errors"
	"strings"
	"testing"

	. "github.com/elves/elvish/cli/apptest"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) LastCmd() (histutil.Entry, error) {
	return histutil.Entry{}, mockError
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

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "foo,bar,baz", Seq: 0})
	Start(f.App, Config{
		Store: store,
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
	f.App.CodeArea().MutateState(func(s *codearea.State) {
		*s = codearea.State{}
	})
	Start(f.App, Config{
		Store: store,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo,bar,baz", term.DotHere)

	// Default wordifier.
	f.App.CodeArea().MutateState(func(s *codearea.State) {
		*s = codearea.State{}
	})
	store.AddCmd(histutil.Entry{Text: "foo bar baz", Seq: 1})
	Start(f.App, Config{Store: store})
	f.TTY.Inject(term.K('0'))
	f.TestTTY(t, "foo", term.DotHere)
}

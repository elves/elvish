package modes

import (
	"strings"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

func TestNewLastcmd_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewLastcmd(f.App, LastcmdSpec{})
	if err != errNoHistoryStore {
		t.Error("expect errNoHistoryStore")
	}
}

func TestNewLastcmd_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		st := histutil.NewMemStore("foo")
		_, err := NewLastcmd(app, LastcmdSpec{Store: st})
		return err
	})
}

func TestNewLastcmd_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	db := histutil.NewFaultyInMemoryDB()
	store, err := histutil.NewDBStore(db)
	if err != nil {
		panic(err)
	}
	db.SetOneOffError(errMock)

	_, err = NewLastcmd(f.App, LastcmdSpec{Store: store})
	if err.Error() != "db error: mock error" {
		t.Error("expect db error")
	}
}

func TestLastcmd(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemStore("foo,bar,baz")
	startLastcmd(f.App, LastcmdSpec{
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
	f.App.ActiveWidget().(tk.CodeArea).MutateState(func(s *tk.CodeAreaState) {
		*s = tk.CodeAreaState{}
	})
	startLastcmd(f.App, LastcmdSpec{
		Store: st,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo,bar,baz", term.DotHere)

	// Default wordifier.
	f.App.ActiveWidget().(tk.CodeArea).MutateState(func(s *tk.CodeAreaState) {
		*s = tk.CodeAreaState{}
	})
	st.AddCmd(storedefs.Cmd{Text: "foo bar baz", Seq: 1})
	startLastcmd(f.App, LastcmdSpec{Store: st})
	f.TTY.Inject(term.K('0'))
	f.TestTTY(t, "foo", term.DotHere)
}

func startLastcmd(app cli.App, spec LastcmdSpec) {
	w, err := NewLastcmd(app, spec)
	startMode(app, w, err)
}

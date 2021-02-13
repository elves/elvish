package mode

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/ui"
)

func TestNewHistlist_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewHistlist(f.App, HistlistSpec{})
	if err != errNoHistoryStore {
		t.Errorf("want errNoHistoryStore")
	}
}

type faultyStore struct{}

var errMock = errors.New("mock error")

func (s faultyStore) AllCmds() ([]store.Cmd, error) { return nil, errMock }

func TestNewHistlist_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewHistlist(f.App, HistlistSpec{AllCmds: faultyStore{}.AllCmds})
	if err.Error() != "db error: mock error" {
		t.Errorf("want db error")
	}
}

func TestHistlist(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemStore(
		// 0    1      2
		"foo", "bar", "baz")
	startHistlist(f.App, HistlistSpec{AllCmds: st.AllCmds})

	// Test initial UI - last item selected
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on)  ", Styles,
		"******************** ", term.DotHere, "\n",
		"   0 foo\n",
		"   1 bar\n",
		"   2 baz                                          ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")

	// Test filtering.
	f.TTY.Inject(term.K('b'))
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on)  b", Styles,
		"********************  ", term.DotHere, "\n",
		"   1 bar\n",
		"   2 baz                                          ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")

	// Test accepting.
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "baz", term.DotHere)

	// Test accepting when there is already some text.
	st.AddCmd(store.Cmd{Text: "baz2"})
	startHistlist(f.App, HistlistSpec{AllCmds: st.AllCmds})
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "baz",
		// codearea now contains newly inserted entry on a separate line
		"\n", "baz2", term.DotHere)
}

func TestHistlist_Dedup(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemStore(
		// 0    1      2
		"ls", "echo", "ls")

	// No dedup
	startHistlist(f.App,
		HistlistSpec{AllCmds: st.AllCmds, Dedup: func() bool { return false }})
	f.TestTTY(t,
		"\n",
		" HISTORY  ", Styles,
		"********* ", term.DotHere, "\n",
		"   0 ls\n",
		"   1 echo\n",
		"   2 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")

	// With dedup
	startHistlist(f.App,
		HistlistSpec{AllCmds: st.AllCmds, Dedup: func() bool { return true }})
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on)  ", Styles,
		"******************** ", term.DotHere, "\n",
		"   1 echo\n",
		"   2 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func TestHistlist_CaseSensitive(t *testing.T) {
	f := Setup(WithTTY(func(tty TTYCtrl) { tty.SetSize(50, 50) }))
	defer f.Stop()

	st := histutil.NewMemStore(
		// 0  1
		"ls", "LS")

	// Case sensitive
	startHistlist(f.App,
		HistlistSpec{AllCmds: st.AllCmds, CaseSensitive: func() bool { return true }})
	f.TTY.Inject(term.K('l'))
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on)  l", Styles,
		"********************  ", term.DotHere, "\n",
		"   0 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")

	// Case insensitive
	startHistlist(f.App,
		HistlistSpec{AllCmds: st.AllCmds, CaseSensitive: func() bool { return false }})
	f.TTY.Inject(term.K('l'))
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on) (case-insensitive)  l", Styles,
		"***************************************  ", term.DotHere, "\n",
		"   0 ls\n",
		"   1 LS                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func startHistlist(app cli.App, spec HistlistSpec) {
	w, err := NewHistlist(app, spec)
	if w != nil {
		app.SetAddon(w, false)
		app.Redraw()
	}
	if err != nil {
		app.Notify(err.Error())
	}
}

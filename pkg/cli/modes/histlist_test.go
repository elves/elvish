package modes

import (
	"regexp"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/store/storedefs"
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

func TestNewHistlist_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		st := histutil.NewMemStore("foo")
		_, err := NewHistlist(app, HistlistSpec{AllCmds: st.AllCmds})
		return err
	})
}

type faultyStore struct{}

func (s faultyStore) AllCmds() ([]storedefs.Cmd, error) { return nil, errMock }

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
	st.AddCmd(storedefs.Cmd{Text: "baz2"})
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

	f.App.PopAddon()

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

func TestHistlist_CustomFilter(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemStore(
		// 0   1         2
		"vi", "elvish", "nvi")

	startHistlist(f.App, HistlistSpec{
		AllCmds: st.AllCmds,
		Filter: FilterSpec{
			Maker: func(p string) func(string) bool {
				re, _ := regexp.Compile(p)
				return func(s string) bool {
					return re != nil && re.MatchString(s)
				}
			},
			Highlighter: func(p string) (ui.Text, []ui.Text) {
				return ui.T(p, ui.Inverse), nil
			},
		},
	})
	f.TTY.Inject(term.K('v'), term.K('i'), term.K('$'))
	f.TestTTY(t,
		"\n",
		" HISTORY (dedup on)  vi$", Styles,
		"******************** +++", term.DotHere, "\n",
		"   0 vi\n",
		"   2 nvi                                          ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func startHistlist(app cli.App, spec HistlistSpec) {
	w, err := NewHistlist(app, spec)
	startMode(app, w, err)
}

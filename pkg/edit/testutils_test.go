package edit

import (
	"fmt"
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

// Aliases.
var (
	Args   = tt.Args
	Styles = clitest.Styles
)

type fixture struct {
	Editor  *Editor
	TTYCtrl clitest.TTYCtrl
	Evaler  *eval.Evaler
	Store   storedefs.Store
	Home    string

	width  int
	codeCh <-chan string
	errCh  <-chan error
}

func rc(codes ...string) func(*fixture) {
	return func(f *fixture) { evals(f.Evaler, codes...) }
}

func assign(name string, val any) func(*fixture) {
	return func(f *fixture) {
		f.Evaler.ExtendGlobal(eval.BuildNs().AddVar("temp", vars.NewReadOnly(val)))
		evals(f.Evaler, "set "+name+" = $temp")
	}
}

func storeOp(storeFn func(storedefs.Store)) func(*fixture) {
	return func(f *fixture) {
		storeFn(f.Store)
		// TODO(xiaq): Don't depend on this Elvish API.
		evals(f.Evaler, "edit:history:fast-forward")
	}
}

func setup(c testutil.Cleanuper, fns ...func(*fixture)) *fixture {
	st := store.MustTempStore(c)
	home := testutil.InTempHome(c)
	testutil.Setenv(c, "PATH", "")

	tty, ttyCtrl := clitest.NewFakeTTY()
	ev := eval.NewEvaler()
	ev.ExtendGlobal(eval.BuildNs().AddNs("file", file.Ns))
	ed := NewEditor(tty, ev, st)
	ev.ExtendBuiltin(eval.BuildNs().AddNs("edit", ed))
	evals(ev,
		// This is the same as the default prompt for non-root users. This makes
		// sure that the tests will work when run as root.
		"set edit:prompt = { tilde-abbr $pwd; put '> ' }",
		// This will simplify most tests against the terminal.
		"set edit:rprompt = { }")
	f := &fixture{Editor: ed, TTYCtrl: ttyCtrl, Evaler: ev, Store: st, Home: home}
	for _, fn := range fns {
		fn(f)
	}
	_, f.width = tty.Size()
	f.codeCh, f.errCh = clitest.StartReadCode(f.Editor.ReadCode)
	c.Cleanup(func() {
		f.Editor.app.CommitEOF()
		f.Wait()
	})
	return f
}

func (f *fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

func (f *fixture) MakeBuffer(args ...any) *term.Buffer {
	return term.NewBufferBuilder(f.width).MarkLines(args...).Buffer()
}

func (f *fixture) TestTTY(t *testing.T, args ...any) {
	t.Helper()
	f.TTYCtrl.TestBuffer(t, f.MakeBuffer(args...))
}

func (f *fixture) TestTTYNotes(t *testing.T, args ...any) {
	t.Helper()
	f.TTYCtrl.TestNotesBuffer(t, f.MakeBuffer(args...))
}

func (f *fixture) SetCodeBuffer(b tk.CodeBuffer) {
	codeArea(f.Editor.app).MutateState(func(s *tk.CodeAreaState) {
		s.Buffer = b
	})
}

func feedInput(ttyCtrl clitest.TTYCtrl, s string) {
	for _, r := range s {
		ttyCtrl.Inject(term.K(r))
	}
}

func evals(ev *eval.Evaler, codes ...string) {
	for _, code := range codes {
		err := ev.Eval(parse.Source{Name: "[test]", Code: code}, eval.EvalCfg{})
		if err != nil {
			panic(fmt.Errorf("eval %q: %s", code, err))
		}
	}
}

func getGlobal(ev *eval.Evaler, name string) any {
	v, _ := ev.Global().Index(name)
	return v
}

func testGlobals(t *testing.T, ev *eval.Evaler, wantVals map[string]any) {
	t.Helper()
	for name, wantVal := range wantVals {
		testGlobal(t, ev, name, wantVal)
	}
}

func testGlobal(t *testing.T, ev *eval.Evaler, name string, wantVal any) {
	t.Helper()
	if val := getGlobal(ev, name); !vals.Equal(val, wantVal) {
		t.Errorf("$%s = %s, want %s",
			name, vals.ReprPlain(val), vals.ReprPlain(wantVal))
	}
}

func testThatOutputErrorIsBubbled(t *testing.T, f *fixture, code string) {
	t.Helper()
	evals(f.Evaler, "var ret = (bool ?("+code+" >&-))")
	// Exceptions are booleanly false
	testGlobal(t, f.Evaler, "ret", false)
}

func codeArea(app cli.App) tk.CodeArea { return app.ActiveWidget().(tk.CodeArea) }

package edit

import (
	"fmt"
	"testing"

	"src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/testutil"
)

var Styles = clitest.Styles

type fixture struct {
	Editor  *Editor
	TTYCtrl clitest.TTYCtrl
	Evaler  *eval.Evaler
	Store   store.Store
	Home    string

	width   int
	codeCh  <-chan string
	errCh   <-chan error
	cleanup func()
}

func rc(codes ...string) func(*fixture) {
	return func(f *fixture) { evals(f.Evaler, codes...) }
}

func assign(name string, val interface{}) func(*fixture) {
	return func(f *fixture) {
		f.Evaler.AddGlobal(eval.CombineNs(f.Evaler.Global(),
			eval.NsBuilder{"temp": vars.NewReadOnly(val)}.Ns()))
		evals(f.Evaler, name+` = $temp`)
	}
}

func storeOp(storeFn func(store.Store)) func(*fixture) {
	return func(f *fixture) {
		storeFn(f.Store)
		// TODO(xiaq): Don't depend on this Elvish API.
		evals(f.Evaler, "edit:history:fast-forward")
	}
}

func setup(fns ...func(*fixture)) *fixture {
	st, cleanupStore := store.MustGetTempStore()
	home, cleanupFs := testutil.InTempHome()
	restorePATH := testutil.WithTempEnv("PATH", "")

	tty, ttyCtrl := clitest.NewFakeTTY()
	ev := eval.NewEvaler()
	ed := NewEditor(tty, ev, st)
	ev.AddModule("edit", ed.Ns())
	evals(ev,
		`use edit`,
		// This is the same as the default prompt for non-root users. This makes
		// sure that the tests will work when run as root.
		"edit:prompt = { tilde-abbr $pwd; put '> ' }",
		// This will simplify most tests against the terminal.
		"edit:rprompt = { }")
	f := &fixture{Editor: ed, TTYCtrl: ttyCtrl, Evaler: ev, Store: st, Home: home}
	for _, fn := range fns {
		fn(f)
	}
	_, f.width = tty.Size()
	f.codeCh, f.errCh = clitest.StartReadCode(f.Editor.ReadCode)
	f.cleanup = func() {
		f.Editor.app.CommitEOF()
		f.Wait()
		restorePATH()
		cleanupFs()
		cleanupStore()
	}
	return f
}

func (f *fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

func (f *fixture) Cleanup() {
	f.cleanup()
}

func (f *fixture) MakeBuffer(args ...interface{}) *term.Buffer {
	return term.NewBufferBuilder(f.width).MarkLines(args...).Buffer()
}

func (f *fixture) TestTTY(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTYCtrl.TestBuffer(t, f.MakeBuffer(args...))
}

func (f *fixture) TestTTYNotes(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTYCtrl.TestNotesBuffer(t, f.MakeBuffer(args...))
}

func (f *fixture) SetCodeBuffer(b tk.CodeBuffer) {
	f.Editor.app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
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

func getGlobal(ev *eval.Evaler, name string) interface{} {
	v, _ := ev.Global().Index(name)
	return v
}

func testGlobals(t *testing.T, ev *eval.Evaler, wantVals map[string]interface{}) {
	t.Helper()
	for name, wantVal := range wantVals {
		testGlobal(t, ev, name, wantVal)
	}
}

func testGlobal(t *testing.T, ev *eval.Evaler, name string, wantVal interface{}) {
	t.Helper()
	if val := getGlobal(ev, name); !vals.Equal(val, wantVal) {
		t.Errorf("$%s = %s, want %s",
			name, vals.Repr(val, vals.NoPretty), vals.Repr(wantVal, vals.NoPretty))
	}
}

package cliedit

import (
	"fmt"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/ui"
)

var styles = map[rune]ui.Styling{
	'-': ui.Underlined,
	'm': ui.JoinStylings(ui.Bold, ui.LightGray, ui.BgMagenta), // mode line
	'#': ui.Inverse,
	'g': ui.Green,                                 // good
	'G': ui.JoinStylings(ui.Green, ui.Underlined), // good with underline
	'b': ui.Red,                                   // bad
	'v': ui.Magenta,                               // variables
	'e': ui.BgRed,                                 // error
}

const (
	testTTYHeight = 24
	testTTYWidth  = 60
)

func bb() *term.BufferBuilder { return term.NewBufferBuilder(testTTYWidth) }

type fixture struct {
	Editor  *Editor
	TTYCtrl cli.TTYCtrl
	Evaler  *eval.Evaler
	Store   storedefs.Store
	Home    string

	codeCh   <-chan string
	errCh    <-chan error
	cleanups []func()
}

func rc(codes ...string) func(*fixture) {
	return func(f *fixture) { evals(f.Evaler, codes...) }
}

func assign(name string, val interface{}) func(*fixture) {
	return func(f *fixture) {
		f.Evaler.Global["temp"] = vars.NewReadOnly(val)
		evals(f.Evaler, name+` = $temp`)
	}
}

func storeOp(storeFn func(storedefs.Store)) func(*fixture) {
	return func(f *fixture) {
		storeFn(f.Store)
		// TODO(xiaq): Don't depend on this Elvish API.
		evals(f.Evaler, "edit:history:fast-forward")
	}
}

func setup(beforeStart ...func(*fixture)) *fixture {
	st, cleanupStore := store.MustGetTempStore()
	home, cleanupFs := eval.InTempHome()
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(testTTYHeight, testTTYWidth)
	ev := eval.NewEvaler()
	ed := NewEditor(tty, ev, st)
	ev.InstallModule("edit", ed.Ns())
	evals(ev,
		`use edit`,
		// This will simplify most tests against the terminal.
		"edit:rprompt = { }")
	f := &fixture{
		Editor: ed, TTYCtrl: ttyCtrl, Evaler: ev, Store: st, Home: home,
		cleanups: []func(){cleanupStore, cleanupFs}}
	for _, fn := range beforeStart {
		fn(f)
	}
	f.Start()
	return f
}

func (f *fixture) Start() {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := f.Editor.ReadCode()
		// Write to the channels and close them. This means that the first read
		// from those channels will get the return value, and subsequent reads
		// will get the zero value of string and error. This means that the Wait
		// method can be called multiple times, and only the first call blocks
		// until the editor stops, and subsequent calls are no-ops.
		codeCh <- code
		close(codeCh)
		errCh <- err
		close(errCh)
	}()
	f.codeCh, f.errCh = codeCh, errCh
	f.cleanups = append(f.cleanups, func() { f.StopAndWait() })
}

func (f *fixture) StopAndWait() (string, error) {
	f.Editor.app.CommitEOF()
	return f.Wait()
}

func (f *fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

func (f *fixture) Cleanup() {
	for i := len(f.cleanups) - 1; i >= 0; i-- {
		f.cleanups[i]()
	}
}

func feedInput(ttyCtrl cli.TTYCtrl, s string) {
	for _, r := range s {
		ttyCtrl.Inject(term.K(r))
	}
}

func evals(ev *eval.Evaler, codes ...string) {
	for _, code := range codes {
		err := ev.EvalSourceInTTY(eval.NewInteractiveSource(code))
		if err != nil {
			panic(fmt.Errorf("eval %q: %s", code, err))
		}
	}
}

func getGlobal(ev *eval.Evaler, name string) interface{} {
	return ev.Global[name].Get()
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

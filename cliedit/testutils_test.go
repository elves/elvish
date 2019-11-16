package cliedit

import (
	"fmt"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/store/storedefs"
)

var styles = map[rune]string{
	'-': "underlined",
	'm': "bold lightgray bg-magenta", // mode line
	'#': "inverse",
	'g': "green",            // good
	'G': "green underlined", // good with underline
	'b': "red",              // bad
	'v': "magenta",          // variables
	'e': "bg-red",           // error
}

const (
	testTTYHeight = 24
	testTTYWidth  = 60
)

func bb() *ui.BufferBuilder { return ui.NewBufferBuilder(testTTYWidth) }

type setupOpt struct {
	// Don't start the editor.
	Unstarted bool
	// Operation on the store before creating the editor.
	StoreOp func(storedefs.Store)
}

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

func setup() *fixture {
	return setupWithOpt(setupOpt{})
}

func setupWithRC(codes ...string) *fixture {
	f := setupWithOpt(setupOpt{Unstarted: true})
	evals(f.Evaler, codes...)
	f.Start()
	return f
}

func setupWithOpt(opt setupOpt) *fixture {
	st, cleanupStore := store.MustGetTempStore()
	if opt.StoreOp != nil {
		opt.StoreOp(st)
	}
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
	if !opt.Unstarted {
		f.Start()
	}
	return f
}

func (f *fixture) Start() {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := f.Editor.ReadLine()
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

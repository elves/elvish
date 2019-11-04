package cliedit

import (
	"fmt"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
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

func feedInput(ttyCtrl cli.TTYCtrl, s string) {
	for _, r := range s {
		ttyCtrl.Inject(term.K(r))
	}
}

func evalf(ev *eval.Evaler, format string, args ...interface{}) {
	code := fmt.Sprintf(format, args...)
	// TODO: Should use a difference source type
	err := ev.EvalSourceInTTY(eval.NewInteractiveSource(code))
	if err != nil {
		panic(err)
	}
}

func setup() (*Editor, cli.TTYCtrl, *eval.Evaler, func()) {
	// TODO(xiaq): Use an in-memory implementation when that is possible.
	st, cleanupStore := store.MustGetTempStore()
	ed, ttyCtrl, ev, cleanupFs := setupWithStore(st)
	return ed, ttyCtrl, ev, func() {
		cleanupFs()
		cleanupStore()
	}
}

func setupWithStore(st storedefs.Store) (*Editor, cli.TTYCtrl, *eval.Evaler, func()) {
	_, cleanup := eval.InTempHome()
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(testTTYHeight, testTTYWidth)
	ev := eval.NewEvaler()
	ed := NewEditor(tty, ev, st)
	ev.InstallModule("edit", ed.Ns())
	evalf(ev, "use edit")
	evalf(ev, "edit:rprompt = { }")
	return ed, ttyCtrl, ev, cleanup
}

func setupStarted() (*Editor, cli.TTYCtrl, *eval.Evaler, func()) {
	ed, ttyCtrl, ev, cleanup := setup()
	_, _, stop := start(ed)
	return ed, ttyCtrl, ev, func() {
		stop()
		cleanup()
	}
}

func start(ed *Editor) (<-chan string, <-chan error, func()) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := ed.ReadLine()
		// Write to the channels and close them. This means that the first read
		// from those channels will get the return value, and subsequent reads
		// will get the zero value of string and error. This in turn implies
		// that:
		//
		// 1) The caller of start can read the return value from the channel
		//    before it calls the stop callback.
		// 2) As long as the code has reached this point, the read from the stop
		//    callback will not block.
		codeCh <- code
		close(codeCh)
		errCh <- err
		close(errCh)
	}()
	return codeCh, errCh, func() {
		ed.app.CommitEOF()
		<-codeCh
	}
}

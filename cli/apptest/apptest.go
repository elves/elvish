// Package apptest provides utilities for testing cli.App.
package apptest

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

// Common stylesheet.
var Styles = ui.RuneStylesheet{
	'_': ui.Underlined,
	'*': ui.Stylings(ui.Bold, ui.LightGray, ui.BgMagenta),
	'+': ui.Inverse,
	'/': ui.Blue,
	'#': ui.Stylings(ui.Inverse, ui.Blue),
	'!': ui.Red,
	'-': ui.Magenta,
	'X': ui.Stylings(ui.Inverse, ui.Magenta),
}

// Fixture is a test fixture.
type Fixture struct {
	App    cli.App
	TTY    cli.TTYCtrl
	width  int
	codeCh <-chan string
	errCh  <-chan error
}

// Setup sets up a test fixture. It contains an App whose ReadCode method has
// been started asynchronously.
func Setup(fns ...func(*cli.AppSpec, cli.TTYCtrl)) *Fixture {
	tty, ttyCtrl := cli.NewFakeTTY()
	spec := cli.AppSpec{TTY: tty}
	for _, fn := range fns {
		fn(&spec, ttyCtrl)
	}
	app := cli.NewApp(spec)
	codeCh, errCh := StartReadCode(app.ReadCode)
	_, width := tty.Size()
	return &Fixture{app, ttyCtrl, width, codeCh, errCh}
}

// WithSpec takes a function that operates on *cli.AppSpec, and wraps it into a
// form suitable for passing to Setup.
func WithSpec(f func(*cli.AppSpec)) func(*cli.AppSpec, cli.TTYCtrl) {
	return func(spec *cli.AppSpec, _ cli.TTYCtrl) { f(spec) }
}

// WithTTY takes a function that operates on cli.TTYCtrl, and wraps it to a form
// suitable for passing to Setup.
func WithTTY(f func(cli.TTYCtrl)) func(*cli.AppSpec, cli.TTYCtrl) {
	return func(_ *cli.AppSpec, tty cli.TTYCtrl) { f(tty) }
}

// Wait waits for ReaCode to finish, and returns its return values.
func (f *Fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

// Stop stops ReadCode and waits for it to finish. If ReadCode has already been
// stopped, it is a no-op.
func (f *Fixture) Stop() {
	f.App.CommitEOF()
	f.Wait()
}

// MakeBuffer is a helper for building a buffer. It is equivalent to
// term.NewBufferBuilder(width of terminal).MarkLines(args...).Buffer().
func (f *Fixture) MakeBuffer(args ...interface{}) *term.Buffer {
	return term.NewBufferBuilder(f.width).MarkLines(args...).Buffer()
}

// TestTTY is equivalent to f.TTY.TestBuffer(f.MakeBuffer(args...)).
func (f *Fixture) TestTTY(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTY.TestBuffer(t, f.MakeBuffer(args...))
}

// TestTTYNotes is equivalent to f.TTY.TestNotesBuffer(f.MakeBuffer(args...)).
func (f *Fixture) TestTTYNotes(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTY.TestNotesBuffer(t, f.MakeBuffer(args...))
}

// StartReadCode starts the given function asynchronously. It returns two
// channels; when the function returns, the return values will be delivered on
// those two channels and the two channels will be closed.
func StartReadCode(readCode func() (string, error)) (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := readCode()
		codeCh <- code
		errCh <- err
		close(codeCh)
		close(errCh)
	}()
	return codeCh, errCh
}

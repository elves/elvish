// Package clitest provides utilities for testing cli.App.
package clitest

import (
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// Styles defines a common stylesheet for unit tests.
var Styles = ui.RuneStylesheet{
	'_': ui.Underlined,
	'b': ui.Bold,
	'*': ui.Stylings(ui.Bold, ui.FgWhite, ui.BgMagenta),
	'+': ui.Inverse,
	'/': ui.FgBlue,
	'#': ui.Stylings(ui.Inverse, ui.FgBlue),
	'!': ui.FgRed,
	'?': ui.Stylings(ui.FgBrightWhite, ui.BgRed),
	'-': ui.FgMagenta,
	'X': ui.Stylings(ui.Inverse, ui.FgMagenta),
	'v': ui.FgGreen,
	'V': ui.Stylings(ui.Underlined, ui.FgGreen),
	'$': ui.FgMagenta,
	'c': ui.FgCyan, // mnemonic "Comment"
}

// Fixture is a test fixture.
type Fixture struct {
	App    cli.App
	TTY    TTYCtrl
	width  int
	codeCh <-chan string
	errCh  <-chan error
}

// Setup sets up a test fixture. It contains an App whose ReadCode method has
// been started asynchronously.
func Setup(fns ...func(*cli.AppSpec, TTYCtrl)) *Fixture {
	tty, ttyCtrl := NewFakeTTY()
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
func WithSpec(f func(*cli.AppSpec)) func(*cli.AppSpec, TTYCtrl) {
	return func(spec *cli.AppSpec, _ TTYCtrl) { f(spec) }
}

// WithTTY takes a function that operates on TTYCtrl, and wraps it to a form
// suitable for passing to Setup.
func WithTTY(f func(TTYCtrl)) func(*cli.AppSpec, TTYCtrl) {
	return func(_ *cli.AppSpec, tty TTYCtrl) { f(tty) }
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
func (f *Fixture) MakeBuffer(args ...any) *term.Buffer {
	return term.NewBufferBuilder(f.width).MarkLines(args...).Buffer()
}

// TestTTY is equivalent to f.TTY.TestBuffer(f.MakeBuffer(args...)).
func (f *Fixture) TestTTY(t *testing.T, args ...any) {
	t.Helper()
	f.TTY.TestBuffer(t, f.MakeBuffer(args...))
}

// TestTTYNotes is equivalent to f.TTY.TestNotesBuffer(f.MakeBuffer(args...)).
func (f *Fixture) TestTTYNotes(t *testing.T, args ...any) {
	t.Helper()
	f.TTY.TestNotesBuffer(t, f.MakeBuffer(args...))
}

// StartReadCode starts the readCode function asynchronously, and returns two
// channels that deliver its return values. The two channels are closed after
// return values are delivered, so that subsequent reads will return zero values
// and not block.
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

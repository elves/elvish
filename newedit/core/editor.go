package core

import (
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/loop"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/sys"
)

// Editor keeps all the state of an interactive line editor throughout its life
// time as well as its dependencies.
type Editor struct {
	loop *loop.Loop
	// Internal dependencies
	render renderCb
	// External dependencies
	tty  TTY
	sigs SignalSource

	Config types.Config
	State  types.State
}

// NewEditor creates a new editor from its two dependencies. The creation does
// not have any observable side effect; a newly created editor is not
// immediately active.
func NewEditor(t TTY, sigs SignalSource) *Editor {
	lp := loop.New()
	ed := &Editor{loop: lp, render: render, tty: t, sigs: sigs}
	lp.HandleCb(ed.handle)
	lp.RedrawCb(ed.redraw)
	return ed
}

func (ed *Editor) handle(e loop.Event) (string, bool) {
	switch e := e.(type) {
	case os.Signal:
		switch e {
		case syscall.SIGHUP:
			return "", true
		case syscall.SIGINT:
			ed.State.Reset()
			ed.Config.TriggerPrompts(true)
		case sys.SIGWINCH:
			ed.Redraw(true)
		}
		return "", false
	case tty.Event:
		switch e := e.(type) {
		case tty.KeyEvent:
			action := getMode(ed.State.Mode()).HandleKey(ui.Key(e), &ed.State)

			switch action {
			case types.CommitCode:
				return ed.State.Code(), true
			}
			ed.Config.TriggerPrompts(false)
		}
		return "", false
	default:
		panic("unreachable")
	}
}

func (ed *Editor) redraw(flag loop.RedrawFlag) {
	var rawState *types.RawState
	final := flag&loop.FinalRedraw != 0
	if final {
		rawState = ed.State.Finalize()
	} else {
		rawState = ed.State.PopForRedraw()
	}

	height, width := ed.tty.Size()
	setup := makeRenderSetup(&ed.Config, height, width)

	bufNotes, bufMain := ed.render(rawState, setup)

	ed.tty.UpdateBuffer(bufNotes, bufMain, flag&loop.FullRedraw != 0)
	if final {
		ed.tty.Newline()
		ed.tty.ResetBuffer()
	}
}

// ReadCode requests the Editor to read code from the terminal. It causes the
// Editor to read events from the terminal and signal source supplied at
// creation, redraws the editor to the terminal on such events, and eventually
// return when an event triggers the current mode to request an exit.
//
// This function is not re-entrant; when it is being executed, the editor is
// said to be active.
func (ed *Editor) ReadCode() (string, error) {
	restore, err := ed.tty.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Relay input events.
	eventCh := ed.tty.StartInput()
	defer ed.tty.StopInput()
	wg.Add(1)
	go func() {
		for event := range eventCh {
			ed.loop.Input(event)
		}
		wg.Done()
	}()

	if ed.sigs != nil {
		// Relay signals.
		sigCh := ed.sigs.NotifySignals()
		defer ed.sigs.StopSignals()
		wg.Add(1)
		go func() {
			for sig := range sigCh {
				ed.loop.Input(sig)
			}
			wg.Done()
		}()
	}

	ed.Config.TriggerPrompts(true)
	// TODO: relay late prompt/rprompt updates.

	// Reset state before returning.
	defer ed.State.Reset()

	// BeforeReadline and AfterReadline hooks.
	for _, f := range ed.Config.BeforeReadline() {
		f()
	}
	defer func() {
		code := ed.State.Code()
		for _, f := range ed.Config.AfterReadline() {
			f(code)
		}
	}()

	return ed.loop.Run()
}

// Like ReadCode, but returns immediately with two channels that will get the
// return values of ReadCode. Useful in tests.
func (ed *Editor) readCodeAsync() (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := ed.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

// Redraw requests a redraw. It never blocks and can be called regardless of
// whether the editor is active or not.
func (ed *Editor) Redraw(full bool) {
	ed.loop.Redraw(full)
}

// Notify adds a note and requests a redraw.
func (ed *Editor) Notify(note string) {
	ed.State.AddNote(note)
	ed.Redraw(false)
}

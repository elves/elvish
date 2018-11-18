package core

import (
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

// Editor keeps all the state of an interactive line editor throughout its life
// time as well as its dependencies.
type Editor struct {
	loop *loop
	// External dependencies
	tty  TTY
	sigs SignalSource

	state types.State

	// Editor configuration that can be modified concurrently.
	Config Config

	// If not nil, will be called when ReadCode starts.
	BeforeReadline func()
	// If not nil, will be called when ReadCode ends; the argument is the code
	// that has been read.
	AfterReadline func(string)

	// Callback for computing the highlight and errors of the code the user has
	// typed.
	Highlighter func(string) (styled.Text, []error)

	// Left-hand and right-hand prompt.
	Prompt, RPrompt Prompt

	// Initial mode.
	InitMode types.Mode
}

// NewEditor creates a new editor from its two dependencies. The creation does
// not have any observable side effect; a newly created editor is not
// immediately active.
func NewEditor(t TTY, sigs SignalSource) *Editor {
	lp := newLoop()
	ed := &Editor{loop: lp, tty: t, sigs: sigs}
	lp.HandleCb(ed.handle)
	lp.RedrawCb(ed.redraw)
	return ed
}

// State returns the editor state.
func (ed *Editor) State() *types.State {
	return &ed.state
}

// A special event type signalling something has seen a late update and a
// refresh is needed. This is currently only used for refreshing prompts.
type lateUpdate struct{}

func (ed *Editor) handle(e event) handleResult {
	switch e := e.(type) {
	case lateUpdate:
		ed.Redraw(false)
		return handleResult{}
	case os.Signal:
		switch e {
		case syscall.SIGHUP:
			return handleResult{quit: true, err: io.EOF}
		case syscall.SIGINT:
			ed.state.Reset()
			ed.triggerPrompts(true)
		case sys.SIGWINCH:
			ed.Redraw(true)
		}
		return handleResult{}
	case tty.Event:
		action := getMode(ed.state.Mode(), ed.InitMode).HandleEvent(e, &ed.state)

		switch action {
		case types.CommitCode:
			return handleResult{quit: true, buffer: ed.state.Code()}
		case types.CommitEOF:
			return handleResult{quit: true, err: io.EOF}
		}
		ed.triggerPrompts(false)
		return handleResult{}
	default:
		panic("unreachable")
	}
}

func (ed *Editor) triggerPrompts(force bool) {
	if ed.Prompt != nil {
		ed.Prompt.Trigger(force)
	}
	if ed.RPrompt != nil {
		ed.RPrompt.Trigger(force)
	}
}

func (ed *Editor) redraw(flag redrawFlag) {
	var rawState *types.RawState
	final := flag&finalRedraw != 0
	if final {
		rawState = ed.state.Finalize()
	} else {
		rawState = ed.state.PopForRedraw()
	}

	height, width := ed.tty.Size()
	if maxHeight := ed.Config.MaxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}
	setup := &renderSetup{
		height, width,
		promptGet(ed.Prompt), promptGet(ed.RPrompt),
		ed.Highlighter, ed.InitMode}

	bufNotes, bufMain := render(rawState, setup)

	ed.tty.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)
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

	// Relay late prompt/rprompt updates.
	stopRelayLateUpdates := make(chan struct{})
	defer close(stopRelayLateUpdates)
	relayLateUpdates := func(p Prompt) {
		if p == nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-p.LateUpdates():
					ed.loop.Input(lateUpdate{})
				case <-stopRelayLateUpdates:
					return
				}
			}
		}()
	}
	relayLateUpdates(ed.Prompt)
	relayLateUpdates(ed.RPrompt)

	// Trigger an initial prompt update.
	ed.triggerPrompts(true)

	// Reset state before returning.
	defer ed.state.Reset()

	// BeforeReadline and AfterReadline hooks.
	if ed.BeforeReadline != nil {
		ed.BeforeReadline()
	}
	if ed.AfterReadline != nil {
		defer func() {
			ed.AfterReadline(ed.state.Code())
		}()
	}

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
	ed.state.AddNote(note)
	ed.Redraw(false)
}

// Callbacks for highlighting code.
type highlighter func(string) (styled.Text, []error)

// Calls the highlighter, falling back to no highlighting if hl is nil.
func (hl highlighter) call(code string) (styled.Text, []error) {
	if hl == nil {
		return styled.Unstyled(code), nil
	}
	return hl(code)
}

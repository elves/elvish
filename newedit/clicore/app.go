package clicore

import (
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/edit/tty"
	clitypes "github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

// App keeps all the state of an CLI throughout its life time as well as its
// dependencies.
type App struct {
	loop *loop
	// External dependencies
	tty  TTY
	sigs SignalSource

	state clitypes.State

	// Configuration that can be modified concurrently.
	Config Config

	// Functions called when ReadCode starts.
	BeforeReadline []func()
	// Functions called when ReadCode ends; the argument is the code that has
	// just been read.
	AfterReadline []func(string)

	// Code highlighter.
	Highlighter Highlighter

	// Left-hand and right-hand prompt.
	Prompt, RPrompt Prompt

	// Initial mode.
	InitMode clitypes.Mode
}

// NewApp creates a new App from its two dependencies. The creation does not
// have any observable side effect; a newly created App is not immediately
// active.
func NewApp(t TTY, sigs SignalSource) *App {
	lp := newLoop()
	app := &App{loop: lp, tty: t, sigs: sigs}
	lp.HandleCb(app.handle)
	lp.RedrawCb(app.redraw)
	return app
}

// State returns the App's state.
func (app *App) State() *clitypes.State {
	return &app.state
}

// A special event type signalling something has seen a late update and a
// refresh is needed. This is currently used for refreshing prompts and
// highlighting.
type lateUpdate struct{}

func (app *App) handle(e event) handleResult {
	switch e := e.(type) {
	case lateUpdate:
		app.Redraw(false)
		return handleResult{}
	case os.Signal:
		switch e {
		case syscall.SIGHUP:
			return handleResult{quit: true, err: io.EOF}
		case syscall.SIGINT:
			app.state.Reset()
			app.triggerPrompts(true)
		case sys.SIGWINCH:
			app.Redraw(true)
		}
		return handleResult{}
	case tty.Event:
		action := getMode(app.state.Mode(), app.InitMode).HandleEvent(e, &app.state)

		switch action {
		case clitypes.CommitCode:
			return handleResult{quit: true, buffer: app.state.Code()}
		case clitypes.CommitEOF:
			return handleResult{quit: true, err: io.EOF}
		}
		app.triggerPrompts(false)
		return handleResult{}
	default:
		panic("unreachable")
	}
}

func (app *App) triggerPrompts(force bool) {
	if app.Prompt != nil {
		app.Prompt.Trigger(force)
	}
	if app.RPrompt != nil {
		app.RPrompt.Trigger(force)
	}
}

var transformerForPending = "underline"

func (app *App) redraw(flag redrawFlag) {
	// Get the state, depending on whether this is the final redraw.
	var rawState *clitypes.RawState
	final := flag&finalRedraw != 0
	if final {
		rawState = app.state.Finalize()
	} else {
		rawState = app.state.PopForRedraw()
	}

	// Get the dimensions available.
	height, width := app.tty.Size()
	if maxHeight := app.Config.MaxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}

	// Prepare the code: applying pending, and highlight.
	code, dot := applyPending(rawState)
	styledCode, errors := highlighterGet(app.Highlighter, code)
	// TODO: Apply transformerForPending to pending code.

	// Render onto buffers.
	setup := &renderSetup{
		height, width,
		promptGet(app.Prompt), promptGet(app.RPrompt),
		styledCode, dot, errors,
		rawState.Notes,
		getMode(rawState.Mode, app.InitMode)}

	bufNotes, bufMain := render(setup)

	// Apply buffers.
	app.tty.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)

	if final {
		app.tty.Newline()
		app.tty.ResetBuffer()
	}
}

func applyPending(st *clitypes.RawState) (code string, dot int) {
	code, dot, pending := st.Code, st.Dot, st.Pending
	if pending != nil {
		code = code[:pending.Begin] + pending.Text + code[pending.End:]
		if dot >= pending.End {
			dot = pending.Begin + len(pending.Text) + (dot - pending.End)
		} else if dot >= pending.Begin {
			dot = pending.Begin + len(pending.Text)
		}
	}
	return code, dot
}

// ReadCode requests the App to read code from the terminal. It causes the App
// to read events from the terminal and signal source supplied at creation,
// redraws to the terminal on such events, and eventually return when an event
// triggers the current mode to request an exit.
//
// This function is not re-entrant; when it is being executed, the App is said
// to be active.
func (app *App) ReadCode() (string, error) {
	restore, err := app.tty.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Relay input events.
	eventCh := app.tty.StartInput()
	defer app.tty.StopInput()
	wg.Add(1)
	go func() {
		for event := range eventCh {
			app.loop.Input(event)
		}
		wg.Done()
	}()

	if app.sigs != nil {
		// Relay signals.
		sigCh := app.sigs.NotifySignals()
		defer app.sigs.StopSignals()
		wg.Add(1)
		go func() {
			for sig := range sigCh {
				app.loop.Input(sig)
			}
			wg.Done()
		}()
	}

	// Relay late updates from prompt, rprompt and highlighter.
	stopRelayLateUpdates := make(chan struct{})
	defer close(stopRelayLateUpdates)
	relayLateUpdates := func(ch <-chan styled.Text) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ch:
					app.loop.Input(lateUpdate{})
				case <-stopRelayLateUpdates:
					return
				}
			}
		}()
	}
	if app.Prompt != nil {
		relayLateUpdates(app.Prompt.LateUpdates())
	}
	if app.RPrompt != nil {
		relayLateUpdates(app.RPrompt.LateUpdates())
	}
	if app.Highlighter != nil {
		relayLateUpdates(app.Highlighter.LateUpdates())
	}

	// Trigger an initial prompt update.
	app.triggerPrompts(true)

	// Reset state before returning.
	defer app.state.Reset()

	// BeforeReadline and AfterReadline hooks.
	for _, f := range app.BeforeReadline {
		f()
	}
	if len(app.AfterReadline) > 0 {
		defer func() {
			for _, f := range app.AfterReadline {
				f(app.state.Code())
			}
		}()
	}

	return app.loop.Run()
}

// Like ReadCode, but returns immediately with two channels that will get the
// return values of ReadCode. Useful in tests.
func (app *App) readCodeAsync() (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := app.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

// Redraw requests a redraw. It never blocks and can be called regardless of
// whether the App is active or not.
func (app *App) Redraw(full bool) {
	app.loop.Redraw(full)
}

// Notify adds a note and requests a redraw.
func (app *App) Notify(note string) {
	app.state.AddNote(note)
	app.Redraw(false)
}

// AddBeforeReadline adds a new before-readline hook function.
func (app *App) AddBeforeReadline(f func()) {
	app.BeforeReadline = append(app.BeforeReadline, f)
}

// AddAfterReadline adds a new after-readline hook function.
func (app *App) AddAfterReadline(f func(string)) {
	app.AfterReadline = append(app.AfterReadline, f)
}

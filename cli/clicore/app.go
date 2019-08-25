package clicore

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/codearea"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

// App keeps all the state of an CLI throughout its life time as well as its
// dependencies.
type App struct {
	loop *loop
	tty  TTY

	StateMutex sync.RWMutex
	State      State

	CodeArea codearea.Widget

	// Configuration.
	Config Config
}

// State represents mutable state of an App.
type State struct {
	// Notes that have been added since the last redraw.
	Notes []string
	// A widget to show under the codearea widget.
	Listing clitypes.Widget
}

// Note appends a new note.
func (s *State) Note(note string) {
	s.Notes = append(s.Notes, note)
}

// Notef is equivalent to calling Note with fmt.Sprintf(format, a...).
func (s *State) Notef(format string, a ...interface{}) {
	s.Note(fmt.Sprintf(format, a...))
}

// PopNotes returns s.Notes and resets s.Notes to an empty slice.
func (s *State) PopNotes() []string {
	notes := s.Notes
	s.Notes = nil
	return notes
}

// NewApp creates a new App from two abstract dependencies. The creation does
// not have any observable side effect; a newly created App is not immediately
// active. This is the most general way to create an App.
func NewApp(t TTY) *App {
	lp := newLoop()
	app := &App{loop: lp, tty: t}
	lp.HandleCb(app.handle)
	lp.RedrawCb(app.redraw)
	return app
}

// MutateAppState calls the given function while locking the state mutex.
func (app *App) MutateAppState(f func(*State)) {
	app.StateMutex.Lock()
	defer app.StateMutex.Unlock()
	f(&app.State)
}

// CopyAppState returns a copy of the app state.
func (app *App) CopyAppState() State {
	app.StateMutex.RLock()
	defer app.StateMutex.RUnlock()
	return app.State
}

func (app *App) resetAllStates() {
	app.MutateAppState(func(s *State) { *s = State{} })
	app.CodeArea.MutateCodeAreaState(
		func(s *codearea.State) { *s = codearea.State{} })
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
			app.resetAllStates()
			app.triggerPrompts(true)
		case sys.SIGWINCH:
			app.Redraw(true)
		}
		return handleResult{}
	case term.Event:
		if listing := app.CopyAppState().Listing; listing != nil {
			listing.Handle(e)
		} else {
			app.CodeArea.Handle(e)
		}

		// TODO(xiaq): Use some kind of return value from the handler instead of
		// hardcoding event.
		switch e {
		case term.K(ui.Enter): // commit code
			buffer := app.CodeArea.CopyState().CodeBuffer.Content
			return handleResult{quit: true, buffer: buffer}
		case term.K('D', ui.Ctrl):
			return handleResult{quit: true, err: io.EOF}
		}

		app.triggerPrompts(false)
		return handleResult{}
	default:
		panic("unreachable")
	}
}

func (app *App) triggerPrompts(force bool) {
	prompt := app.Config.Prompt
	rprompt := app.Config.RPrompt
	if prompt != nil {
		prompt.Trigger(force)
	}
	if rprompt != nil {
		rprompt.Trigger(force)
	}
}

var transformerForPending = "underline"

func (app *App) redraw(flag redrawFlag) {
	// Get the dimensions available.
	height, width := app.tty.Size()
	if maxHeight := app.Config.maxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}

	var notes []string
	var listing clitypes.Renderer
	app.MutateAppState(func(s *State) {
		notes = s.PopNotes()
		listing = s.Listing
	})

	bufNotes := renderNotes(notes, width)
	bufMain := mainRenderer{&app.CodeArea, listing}.Render(width, height)

	// Apply buffers.
	app.tty.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)

	if flag&finalRedraw != 0 {
		app.tty.Newline()
		app.tty.ResetBuffer()
	}
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

	// Relay signals.
	sigCh := app.tty.NotifySignals()
	defer app.tty.StopSignals()
	wg.Add(1)
	go func() {
		for sig := range sigCh {
			app.loop.Input(sig)
		}
		wg.Done()
	}()

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
	if prompt := app.Config.Prompt; prompt != nil {
		app.CodeArea.Prompt = prompt.Get
		relayLateUpdates(prompt.LateUpdates())
	}
	if rprompt := app.Config.RPrompt; rprompt != nil {
		app.CodeArea.RPrompt = rprompt.Get
		relayLateUpdates(rprompt.LateUpdates())
	}
	if highlighter := app.Config.Highlighter; highlighter != nil {
		app.CodeArea.Highlighter = highlighter.Get
		relayLateUpdates(highlighter.LateUpdates())
	}

	// Trigger an initial prompt update.
	app.triggerPrompts(true)

	// Reset state before returning.
	defer app.resetAllStates()

	// BeforeReadline and AfterReadline hooks.
	app.Config.beforeReadline()
	defer func() {
		app.Config.afterReadline(app.CodeArea.CopyState().CodeBuffer.Content)
	}()

	return app.loop.Run()
}

// ReadCodeAsync is an asynchronous version of ReadCode. It returns immediately
// with two channels that will get the return values of ReadCode. Mainly useful
// in tests.
func (app *App) ReadCodeAsync() (<-chan string, <-chan error) {
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

// CommitEOF causes the main loop to exit with EOF.
func (app *App) CommitEOF() {
	// TODO: Implement.
}

// CommitCode causes the main loop to exit with the current code content.
func (app *App) CommitCode() {
	// TODO: Implement.
}

// Notify adds a note and requests a redraw.
func (app *App) Notify(note string) {
	app.MutateAppState(func(s *State) { s.Note(note) })
	app.Redraw(false)
}

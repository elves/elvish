package cli

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

// App represents a CLI app.
type App interface {
	// MutateAppState mutates the state of the app.
	MutateAppState(f func(*State))
	// CopyAppState returns a copy of the a state.
	CopyAppState() State
	// CodeArea returns the codearea widget of the app.
	CodeArea() codearea.Widget
	// ReadCode requests the App to read code from the terminal by running an
	// event loop. This function is not re-entrant.
	ReadCode() (string, error)
	// ReadCodeAsync is an asynchronous version of ReadCode. It returns
	// immediately with two channels that will get the return values of
	// ReadCode. Mainly useful in tests.
	ReadCodeAsync() (<-chan string, <-chan error)
	// Redraw requests a redraw. It never blocks and can be called regardless of
	// whether the App is active or not.
	Redraw()
	// CommitEOF causes the main loop to exit with EOF. If this method is called
	// when an event is being handled, the main loop will exit after the handler
	// returns.
	CommitEOF()
	// CommitCode causes the main loop to exit with the given code content. If
	// this method is called when an event is being handled, the main loop will
	// exit after the handler returns.
	CommitCode(code string)
	// Notify adds a note and requests a redraw.
	Notify(note string)
}

type app struct {
	loop *loop

	StateMutex sync.RWMutex
	AppSpec
	codeArea codearea.Widget
}

// State represents mutable state of an App.
type State struct {
	// Notes that have been added since the last redraw.
	Notes []string
	// A widget to show under the codearea widget.
	Listing el.Widget
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

// NewApp creates a new App from the given specification.
func NewApp(spec AppSpec) App {
	var a app
	fixSpec(&spec)
	spec.CodeArea.OnSubmit = a.CommitCode
	lp := newLoop()
	a = app{loop: lp, AppSpec: spec, codeArea: codearea.New(spec.CodeArea)}
	lp.HandleCb(a.handle)
	lp.RedrawCb(a.redraw)
	return &a
}

func (a *app) MutateAppState(f func(*State)) {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	f(&a.State)
}

func (a *app) CopyAppState() State {
	a.StateMutex.RLock()
	defer a.StateMutex.RUnlock()
	return a.State
}

func (a *app) CodeArea() codearea.Widget {
	return a.codeArea
}

func (a *app) resetAllStates() {
	a.MutateAppState(func(s *State) { *s = State{} })
	a.codeArea.MutateCodeAreaState(
		func(s *codearea.State) { *s = codearea.State{} })
}

func (a *app) handle(e event) {
	switch e := e.(type) {
	case os.Signal:
		switch e {
		case syscall.SIGHUP:
			a.loop.Return("", io.EOF)
		case syscall.SIGINT:
			a.resetAllStates()
			a.triggerPrompts(true)
		case sys.SIGWINCH:
			a.RedrawFull()
		}
	case term.Event:
		if listing := a.CopyAppState().Listing; listing != nil {
			listing.Handle(e)
		} else {
			a.codeArea.Handle(e)
		}
		if !a.loop.HasReturned() {
			a.triggerPrompts(false)
		}
	default:
		panic("unreachable")
	}
}

func (a *app) triggerPrompts(force bool) {
	prompt := a.AppSpec.Prompt
	rprompt := a.AppSpec.RPrompt
	if prompt != nil {
		prompt.Trigger(force)
	}
	if rprompt != nil {
		rprompt.Trigger(force)
	}
}

var transformerForPending = "underline"

func (a *app) redraw(flag redrawFlag) {
	// Get the dimensions available.
	height, width := a.TTY.Size()
	if maxHeight := a.AppSpec.MaxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}

	var notes []string
	var listing el.Renderer
	a.MutateAppState(func(s *State) {
		notes = s.PopNotes()
		listing = s.Listing
	})

	bufNotes := renderNotes(notes, width)
	isFinalRedraw := flag&finalRedraw != 0
	if isFinalRedraw {
		// This is a bit of hack to achieve two things desirable for the final
		// redraw: put the cursor below the code area, and make sure it is on a
		// new empty line.
		listing = layout.Empty{}
	}
	bufMain := renderApp(a.codeArea, listing, width, height)

	// Apply buffers.
	a.TTY.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)

	if isFinalRedraw {
		a.TTY.ResetBuffer()
	}
}

// Renders notes. This does not respect height so that overflow notes end up in
// the scrollback buffer.
func renderNotes(notes []string, width int) *ui.Buffer {
	if len(notes) == 0 {
		return nil
	}
	bb := ui.NewBufferBuilder(width)
	for i, note := range notes {
		if i > 0 {
			bb.Newline()
		}
		bb.WritePlain(note)
	}
	return bb.Buffer()
}

// Renders the codearea, and uses the rest of the height for the listing.
func renderApp(codeArea, listing el.Renderer, width, height int) *ui.Buffer {
	buf := codeArea.Render(width, height)
	if listing != nil && len(buf.Lines) < height {
		bufListing := listing.Render(width, height-len(buf.Lines))
		buf.Extend(bufListing, true)
	}
	return buf
}

func (a *app) ReadCode() (string, error) {
	restore, err := a.TTY.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Relay input events.
	eventCh := a.TTY.StartInput()
	defer a.TTY.StopInput()
	wg.Add(1)
	go func() {
		for event := range eventCh {
			a.loop.Input(event)
		}
		wg.Done()
	}()

	// Relay signals.
	sigCh := a.TTY.NotifySignals()
	defer a.TTY.StopSignals()
	wg.Add(1)
	go func() {
		for sig := range sigCh {
			a.loop.Input(sig)
		}
		wg.Done()
	}()

	// Relay late updates from prompt, rprompt and highlighter.
	stopRelayLateUpdates := make(chan struct{})
	defer close(stopRelayLateUpdates)
	relayLateUpdates := func(ch <-chan styled.Text) {
		if ch == nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ch:
					a.Redraw()
				case <-stopRelayLateUpdates:
					return
				}
			}
		}()
	}

	relayLateUpdates(a.Prompt.LateUpdates())
	relayLateUpdates(a.RPrompt.LateUpdates())
	relayLateUpdates(a.Highlighter.LateUpdates())

	// Trigger an initial prompt update.
	a.triggerPrompts(true)

	// Reset state before returning.
	defer a.resetAllStates()

	// BeforeReadline and AfterReadline hooks.
	a.BeforeReadline()
	defer func() {
		a.AfterReadline(a.codeArea.CopyState().CodeBuffer.Content)
	}()

	return a.loop.Run()
}

func (a *app) ReadCodeAsync() (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := a.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

func (a *app) Redraw() {
	a.loop.Redraw(false)
}

func (a *app) RedrawFull() {
	// This is currently not exposed, but can be exposed later if the need arises.
	a.loop.Redraw(true)
}

func (a *app) CommitEOF() {
	a.loop.Return("", io.EOF)
}

func (a *app) CommitCode(code string) {
	a.loop.Return(code, nil)
}

func (a *app) Notify(note string) {
	a.MutateAppState(func(s *State) { s.Note(note) })
	a.Redraw()
}

// Package cli implements a generic interactive line editor.
package cli

import (
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

// App represents a CLI app.
type App interface {
	// MutateState mutates the state of the app.
	MutateState(f func(*State))
	// CopyState returns a copy of the a state.
	CopyState() State
	// CodeArea returns the codearea widget of the app.
	CodeArea() codearea.Widget
	// ReadCode requests the App to read code from the terminal by running an
	// event loop. This function is not re-entrant.
	ReadCode() (string, error)
	// Redraw requests a redraw. It never blocks and can be called regardless of
	// whether the App is active or not.
	Redraw()
	// CommitEOF causes the main loop to exit with EOF. If this method is called
	// when an event is being handled, the main loop will exit after the handler
	// returns.
	CommitEOF()
	// CommitCode causes the main loop to exit with the current code content. If
	// this method is called when an event is being handled, the main loop will
	// exit after the handler returns.
	CommitCode()
	// Notify adds a note and requests a redraw.
	Notify(note string)
}

type app struct {
	loop *loop

	TTY               TTY
	MaxHeight         func() int
	RPromptPersistent func() bool
	BeforeReadline    func()
	AfterReadline     func(string)
	Highlighter       Highlighter
	Prompt            Prompt
	RPrompt           Prompt

	StateMutex sync.RWMutex
	State      State

	codeArea codearea.Widget
}

// State represents mutable state of an App.
type State struct {
	// Notes that have been added since the last redraw.
	Notes []string
	// A widget to show under the codearea widget.
	Listing el.Widget
}

// NewApp creates a new App from the given specification.
func NewApp(spec AppSpec) App {
	lp := newLoop()
	a := app{
		loop:              lp,
		TTY:               spec.TTY,
		MaxHeight:         spec.MaxHeight,
		RPromptPersistent: spec.RPromptPersistent,
		BeforeReadline:    spec.BeforeReadline,
		AfterReadline:     spec.AfterReadline,
		Highlighter:       spec.Highlighter,
		Prompt:            spec.Prompt,
		RPrompt:           spec.RPrompt,
		State:             spec.State,
	}
	if a.TTY == nil {
		a.TTY, _ = NewFakeTTY()
	}
	if a.MaxHeight == nil {
		a.MaxHeight = func() int { return -1 }
	}
	if a.RPromptPersistent == nil {
		a.RPromptPersistent = func() bool { return false }
	}
	if a.BeforeReadline == nil {
		a.BeforeReadline = func() {}
	}
	if a.AfterReadline == nil {
		a.AfterReadline = func(string) {}
	}
	if a.Highlighter == nil {
		a.Highlighter = dummyHighlighter{}
	}
	if a.Prompt == nil {
		a.Prompt = constPrompt{}
	}
	if a.RPrompt == nil {
		a.RPrompt = constPrompt{}
	}
	lp.HandleCb(a.handle)
	lp.RedrawCb(a.redraw)

	a.codeArea = codearea.New(codearea.Spec{
		OverlayHandler: spec.OverlayHandler,
		Highlighter:    a.Highlighter.Get,
		Prompt:         a.Prompt.Get,
		RPrompt:        a.RPrompt.Get,
		Abbreviations:  spec.Abbreviations,
		QuotePaste:     spec.QuotePaste,
		OnSubmit:       a.CommitCode,
		State:          spec.CodeAreaState,
	})

	return &a
}

func (a *app) MutateState(f func(*State)) {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	f(&a.State)
}

func (a *app) CopyState() State {
	a.StateMutex.RLock()
	defer a.StateMutex.RUnlock()
	return a.State
}

func (a *app) CodeArea() codearea.Widget {
	return a.codeArea
}

func (a *app) resetAllStates() {
	a.MutateState(func(s *State) { *s = State{} })
	a.codeArea.MutateState(
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
		if listing := a.CopyState().Listing; listing != nil {
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
	a.Prompt.Trigger(force)
	a.RPrompt.Trigger(force)
}

var transformerForPending = "underline"

func (a *app) redraw(flag redrawFlag) {
	// Get the dimensions available.
	height, width := a.TTY.Size()
	if maxHeight := a.MaxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}

	var notes []string
	var listing el.Renderer
	a.MutateState(func(s *State) {
		notes, listing = s.Notes, s.Listing
		s.Notes = nil
	})

	bufNotes := renderNotes(notes, width)
	isFinalRedraw := flag&finalRedraw != 0
	if isFinalRedraw {
		hideRPrompt := !a.RPromptPersistent()
		if hideRPrompt {
			a.codeArea.MutateState(func(s *codearea.State) { s.HideRPrompt = true })
		}
		bufMain := renderApp(a.codeArea, nil /* listing */, width, height)
		if hideRPrompt {
			a.codeArea.MutateState(func(s *codearea.State) { s.HideRPrompt = false })
		}
		// Insert a newline after the buffer and position the cursor there.
		bufMain.Extend(ui.NewBuffer(width), true)

		a.TTY.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)
		a.TTY.ResetBuffer()
	} else {
		bufMain := renderApp(a.codeArea, listing, width, height)
		a.TTY.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)
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
		a.AfterReadline(a.codeArea.CopyState().Buffer.Content)
	}()

	return a.loop.Run()
}

// ReadCodeAsync is an asynchronous version of App.ReadCode. Instead of
// blocking, it returns immediately with two channels that will deliver the
// return values of ReadCode when ReadCode returns.
//
// This function is mainly useful in tests.
func ReadCodeAsync(a App) (<-chan string, <-chan error) {
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

func (a *app) CommitCode() {
	code := a.codeArea.CopyState().Buffer.Content
	a.loop.Return(code, nil)
}

func (a *app) Notify(note string) {
	a.MutateState(func(s *State) { s.Notes = append(s.Notes, note) })
	a.Redraw()
}

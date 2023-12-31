// Package cli implements a generic interactive line editor.
package cli

import (
	"io"
	"os"
	"sort"
	"sync"
	"syscall"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/sys"
	"src.elv.sh/pkg/ui"
)

// App represents a CLI app.
type App interface {
	// ReadCode requests the App to read code from the terminal by running an
	// event loop. This function is not re-entrant.
	ReadCode() (string, error)

	// MutateState mutates the state of the app.
	MutateState(f func(*State))
	// CopyState returns a copy of the a state.
	CopyState() State

	// PushAddon pushes a widget to the addon stack.
	PushAddon(w tk.Widget)
	// PopAddon pops the last widget from the addon stack. If the widget
	// implements interface{ Dismiss() }, the Dismiss method is called
	// first. This method does nothing if the addon stack is empty.
	PopAddon()

	// ActiveWidget returns the currently active widget. If the addon stack is
	// non-empty, it returns the last addon. Otherwise it returns the main code
	// area widget.
	ActiveWidget() tk.Widget
	// FocusedWidget returns the currently focused widget. It is searched like
	// ActiveWidget, but skips widgets that implement interface{ Focus() bool }
	// and return false when .Focus() is called.
	FocusedWidget() tk.Widget

	// CommitEOF causes the main loop to exit with EOF. If this method is called
	// when an event is being handled, the main loop will exit after the handler
	// returns.
	CommitEOF()
	// CommitCode causes the main loop to exit with the current code content. If
	// this method is called when an event is being handled, the main loop will
	// exit after the handler returns.
	CommitCode()

	// Redraw requests a redraw. It never blocks and can be called regardless of
	// whether the App is active or not.
	Redraw()
	// RedrawFull requests a full redraw. It never blocks and can be called
	// regardless of whether the App is active or not.
	RedrawFull()
	// Notify adds a note and requests a redraw.
	Notify(note ui.Text)
}

type app struct {
	loop    *loop
	reqRead chan struct{}

	TTY               TTY
	MaxHeight         func() int
	RPromptPersistent func() bool
	BeforeReadline    []func()
	AfterReadline     []func(string)
	Highlighter       Highlighter
	Prompt            Prompt
	RPrompt           Prompt
	GlobalBindings    tk.Bindings

	StateMutex sync.RWMutex
	State      State

	codeArea tk.CodeArea
}

// State represents mutable state of an App.
type State struct {
	// Notes that have been added since the last redraw.
	Notes []ui.Text
	// The addon stack. All widgets are shown under the codearea widget. The
	// last widget handles terminal events.
	Addons []tk.Widget
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
		GlobalBindings:    spec.GlobalBindings,
		State:             spec.State,
	}
	if a.TTY == nil {
		a.TTY = NewTTY(os.Stdin, os.Stderr)
	}
	if a.MaxHeight == nil {
		a.MaxHeight = func() int { return -1 }
	}
	if a.RPromptPersistent == nil {
		a.RPromptPersistent = func() bool { return false }
	}
	if a.Highlighter == nil {
		a.Highlighter = dummyHighlighter{}
	}
	if a.Prompt == nil {
		a.Prompt = NewConstPrompt(nil)
	}
	if a.RPrompt == nil {
		a.RPrompt = NewConstPrompt(nil)
	}
	if a.GlobalBindings == nil {
		a.GlobalBindings = tk.DummyBindings{}
	}
	lp.HandleCb(a.handle)
	lp.RedrawCb(a.redraw)

	a.codeArea = tk.NewCodeArea(tk.CodeAreaSpec{
		Bindings:    spec.CodeAreaBindings,
		Highlighter: a.Highlighter.Get,
		Prompt:      a.Prompt.Get,
		RPrompt:     a.RPrompt.Get,
		QuotePaste:  spec.QuotePaste,
		OnSubmit:    a.CommitCode,
		State:       spec.CodeAreaState,

		SimpleAbbreviations:    spec.SimpleAbbreviations,
		CommandAbbreviations:   spec.CommandAbbreviations,
		SmallWordAbbreviations: spec.SmallWordAbbreviations,
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
	return State{
		append([]ui.Text(nil), a.State.Notes...),
		append([]tk.Widget(nil), a.State.Addons...),
	}
}

type dismisser interface {
	Dismiss()
}

func (a *app) PushAddon(w tk.Widget) {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	a.State.Addons = append(a.State.Addons, w)
}

func (a *app) PopAddon() {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	if len(a.State.Addons) == 0 {
		return
	}
	if d, ok := a.State.Addons[len(a.State.Addons)-1].(dismisser); ok {
		d.Dismiss()
	}
	a.State.Addons = a.State.Addons[:len(a.State.Addons)-1]
}

func (a *app) ActiveWidget() tk.Widget {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	if len(a.State.Addons) > 0 {
		return a.State.Addons[len(a.State.Addons)-1]
	}
	return a.codeArea
}

func (a *app) FocusedWidget() tk.Widget {
	a.StateMutex.Lock()
	defer a.StateMutex.Unlock()
	addons := a.State.Addons
	for i := len(addons) - 1; i >= 0; i-- {
		if hasFocus(addons[i]) {
			return addons[i]
		}
	}
	return a.codeArea
}

func (a *app) resetAllStates() {
	a.MutateState(func(s *State) { *s = State{} })
	a.codeArea.MutateState(
		func(s *tk.CodeAreaState) { *s = tk.CodeAreaState{} })
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
		target := a.ActiveWidget()
		handled := target.Handle(e)
		if !handled {
			handled = a.GlobalBindings.Handle(target, e)
		}
		if !handled {
			if k, ok := e.(term.KeyEvent); ok {
				a.Notify(ui.T("Unbound key: " + ui.Key(k).String()))
			}
		}
		if !a.loop.HasReturned() {
			a.triggerPrompts(false)
			a.reqRead <- struct{}{}
		}
	}
}

func (a *app) triggerPrompts(force bool) {
	a.Prompt.Trigger(force)
	a.RPrompt.Trigger(force)
}

func (a *app) redraw(flag redrawFlag) {
	// Get the dimensions available.
	height, width := a.TTY.Size()
	if maxHeight := a.MaxHeight(); maxHeight > 0 && maxHeight < height {
		height = maxHeight
	}

	var notes []ui.Text
	var addons []tk.Widget
	a.MutateState(func(s *State) {
		notes = s.Notes
		s.Notes = nil
		addons = append([]tk.Widget(nil), s.Addons...)
	})

	bufNotes := renderNotes(notes, width)
	isFinalRedraw := flag&finalRedraw != 0
	if isFinalRedraw {
		hideRPrompt := !a.RPromptPersistent()
		a.codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.HideTips = true
			s.HideRPrompt = hideRPrompt
		})
		bufMain := renderApp([]tk.Widget{a.codeArea /* no addon */}, width, height)
		a.codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.HideTips = false
			s.HideRPrompt = false
		})
		// Insert a newline after the buffer and position the cursor there.
		bufMain.Extend(term.NewBuffer(width), true)

		a.TTY.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)
		a.TTY.ResetBuffer()
	} else {
		bufMain := renderApp(append([]tk.Widget{a.codeArea}, addons...), width, height)
		a.TTY.UpdateBuffer(bufNotes, bufMain, flag&fullRedraw != 0)
	}
}

// Renders notes. This does not respect height so that overflow notes end up in
// the scrollback buffer.
func renderNotes(notes []ui.Text, width int) *term.Buffer {
	if len(notes) == 0 {
		return nil
	}
	bb := term.NewBufferBuilder(width)
	for i, note := range notes {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteStyled(note)
	}
	return bb.Buffer()
}

// Renders the codearea, and uses the rest of the height for the listing.
func renderApp(widgets []tk.Widget, width, height int) *term.Buffer {
	heights, focus := distributeHeight(widgets, width, height)
	var buf *term.Buffer
	for i, w := range widgets {
		if heights[i] == 0 {
			continue
		}
		buf2 := w.Render(width, heights[i])
		if buf == nil {
			buf = buf2
		} else {
			buf.Extend(buf2, i == focus)
		}
	}
	return buf
}

// Distributes the height among all the widgets. Returns the height for each
// widget, and the index of the widget currently focused.
func distributeHeight(widgets []tk.Widget, width, height int) ([]int, int) {
	var focus int
	for i, w := range widgets {
		if hasFocus(w) {
			focus = i
		}
	}
	n := len(widgets)
	heights := make([]int, n)
	if height <= n {
		// Not enough (or just enough) height to render every widget with a
		// height of 1.
		remain := height
		// Start from the focused widget, and extend downwards as much as
		// possible.
		for i := focus; i < n && remain > 0; i++ {
			heights[i] = 1
			remain--
		}
		// If there is still space remaining, start from the focused widget
		// again, and extend upwards as much as possible.
		for i := focus - 1; i >= 0 && remain > 0; i-- {
			heights[i] = 1
			remain--
		}
		return heights, focus
	}

	maxHeights := make([]int, n)
	for i, w := range widgets {
		maxHeights[i] = w.MaxHeight(width, height)
	}

	// The algorithm below achieves the following goals:
	//
	// 1. If maxHeights[u] > maxHeights[v], heights[u] >= heights[v];
	//
	// 2. While achieving goal 1, have as many widgets s.t. heights[u] ==
	//    maxHeights[u].
	//
	// This is done by allocating the height among the widgets following an
	// non-decreasing order of maxHeights. At each step:
	//
	// - If it's possible to allocate maxHeights[u] to all remaining widgets,
	//   then allocate maxHeights[u] to widget u;
	//
	// - If not, allocate the remaining budget evenly - rounding down at each
	//   step, so the widgets with smaller maxHeights gets smaller heights.

	// TODO: Add a test for this.

	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return maxHeights[indices[i]] < maxHeights[indices[j]]
	})

	remain := height
	for rank, idx := range indices {
		if remain >= maxHeights[idx]*(n-rank) {
			heights[idx] = maxHeights[idx]
		} else {
			heights[idx] = remain / (n - rank)
		}
		remain -= heights[idx]
	}

	return heights, focus
}

func hasFocus(w any) bool {
	if f, ok := w.(interface{ Focus() bool }); ok {
		return f.Focus()
	}
	return true
}

func (a *app) ReadCode() (string, error) {
	for _, f := range a.BeforeReadline {
		f()
	}
	defer func() {
		content := a.codeArea.CopyState().Buffer.Content
		for _, f := range a.AfterReadline {
			f(content)
		}
		a.resetAllStates()
	}()

	restore, err := a.TTY.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Relay input events.
	a.reqRead = make(chan struct{}, 1)
	a.reqRead <- struct{}{}
	defer close(a.reqRead)
	defer a.TTY.CloseReader()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range a.reqRead {
			event, err := a.TTY.ReadEvent()
			if err == nil {
				a.loop.Input(event)
			} else if err == term.ErrStopped {
				return
			} else if term.IsReadErrorRecoverable(err) {
				a.loop.Input(term.NonfatalErrorEvent{Err: err})
			} else {
				a.loop.Input(term.FatalErrorEvent{Err: err})
				return
			}
		}
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
	relayLateUpdates := func(ch <-chan struct{}) {
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

	return a.loop.Run()
}

func (a *app) Redraw() {
	a.loop.Redraw(false)
}

func (a *app) RedrawFull() {
	a.loop.Redraw(true)
}

func (a *app) CommitEOF() {
	a.loop.Return("", io.EOF)
}

func (a *app) CommitCode() {
	code := a.codeArea.CopyState().Buffer.Content
	a.loop.Return(code, nil)
}

func (a *app) Notify(note ui.Text) {
	a.MutateState(func(s *State) { s.Notes = append(s.Notes, note) })
	a.Redraw()
}

// Package textview implements a widget for displaying text.
package textview

import (
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

// Widget supports displaying text, and scrolling vertically.
//
// NOTE: This widget now always crops long lines. In future it should support
// wrapping and horizontal scrolling.
type Widget struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	// Publically accessible state.
	State State

	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
	// If true, a vertical scrollbar will be shown when there are more lines
	// that can be displayed, and the widget responds to Up and Down keys.
	Scrollable bool
}

// State keeps publically accessible state of the Widget.
type State struct {
	Lines []string
	First int
}

func (w *Widget) init() {
	if w.OverlayHandler == nil {
		w.OverlayHandler = el.DummyHandler{}
	}
}

func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()

	lines, first := w.getStateForRender(height)
	needScrollbar := w.Scrollable && (first > 0 || first+height < len(lines))
	textWidth := width
	if needScrollbar {
		textWidth--
	}

	bb := ui.NewBufferBuilder(textWidth)
	for i := first; i < first+height && i < len(lines); i++ {
		if i > first {
			bb.Newline()
		}
		bb.WritePlain(util.TrimWcwidth(lines[i], textWidth))
	}
	buf := bb.Buffer()

	if needScrollbar {
		scrollbar := layout.VScrollbar{
			Total: len(lines), Low: first, High: first + height}
		buf.ExtendRight(scrollbar.Render(1, height))
	}
	return buf
}

func (w *Widget) getStateForRender(height int) (lines []string, first int) {
	w.MutateState(func(s *State) {
		if s.First > len(s.Lines)-height && len(s.Lines)-height >= 0 {
			s.First = len(s.Lines) - height
		}
		lines, first = s.Lines, s.First
	})
	return
}

func (w *Widget) Handle(event term.Event) bool {
	w.init()

	if w.OverlayHandler.Handle(event) {
		return true
	}

	if w.Scrollable {
		switch event {
		case term.K(ui.Up):
			w.ScrollBy(-1)
			return true
		case term.K(ui.Down):
			w.ScrollBy(1)
			return true
		}
	}
	return false
}

// ScrollBy scrolls the widget by the given delta. Positive values scroll down,
// and negative values scroll up.
func (w *Widget) ScrollBy(delta int) {
	w.MutateState(func(s *State) {
		s.First += delta
		if s.First < 0 {
			s.First = 0
		}
		if s.First >= len(s.Lines) {
			s.First = len(s.Lines) - 1
		}
	})
}

// MutateState calls the given function while locking the StateMutex.
func (w *Widget) MutateState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

// CopyState returns a copy of the State while r-locking the StateMutex.
func (w *Widget) CopyState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

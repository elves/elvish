// Package textview implements a widget for displaying text with vertical
// scrolling.
package textview

import (
	"sync"

	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/layout"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
	"github.com/elves/elvish/pkg/util"
)

// TextView is a Widget for displaying text, with support for vertical
// scrolling.
//
// NOTE: This widget now always crops long lines. In future it should support
// wrapping and horizontal scrolling.
type TextView interface {
	el.Widget
	// ScrollBy scrolls the widget by the given delta. Positive values scroll
	// down, and negative values scroll up.
	ScrollBy(delta int)
	// MutateState mutates the state.
	MutateState(f func(*TextViewState))
	// CopyState returns a copy of the State.
	CopyState() TextViewState
}

// TextViewSpec specifies the configuration and initial state for a Widget.
type TextViewSpec struct {
	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
	// If true, a vertical scrollbar will be shown when there are more lines
	// that can be displayed, and the widget responds to Up and Down keys.
	Scrollable bool
	// State. Specifies the initial state if used in New.
	State TextViewState
}

// TextViewState keeps mutable state of TextView.
type TextViewState struct {
	Lines []string
	First int
}

type textView struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	TextViewSpec
}

// NewTextView builds a TextView from the given spec.
func NewTextView(spec TextViewSpec) TextView {
	if spec.OverlayHandler == nil {
		spec.OverlayHandler = el.DummyHandler{}
	}
	return &textView{TextViewSpec: spec}
}

func (w *textView) Render(width, height int) *term.Buffer {
	lines, first := w.getStateForRender(height)
	needScrollbar := w.Scrollable && (first > 0 || first+height < len(lines))
	textWidth := width
	if needScrollbar {
		textWidth--
	}

	bb := term.NewBufferBuilder(textWidth)
	for i := first; i < first+height && i < len(lines); i++ {
		if i > first {
			bb.Newline()
		}
		bb.Write(util.TrimWcwidth(lines[i], textWidth))
	}
	buf := bb.Buffer()

	if needScrollbar {
		scrollbar := layout.VScrollbar{
			Total: len(lines), Low: first, High: first + height}
		buf.ExtendRight(scrollbar.Render(1, height))
	}
	return buf
}

func (w *textView) getStateForRender(height int) (lines []string, first int) {
	w.MutateState(func(s *TextViewState) {
		if s.First > len(s.Lines)-height && len(s.Lines)-height >= 0 {
			s.First = len(s.Lines) - height
		}
		lines, first = s.Lines, s.First
	})
	return
}

func (w *textView) Handle(event term.Event) bool {
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

func (w *textView) ScrollBy(delta int) {
	w.MutateState(func(s *TextViewState) {
		s.First += delta
		if s.First < 0 {
			s.First = 0
		}
		if s.First >= len(s.Lines) {
			s.First = len(s.Lines) - 1
		}
	})
}

func (w *textView) MutateState(f func(*TextViewState)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

// CopyState returns a copy of the State while r-locking the StateMutex.
func (w *textView) CopyState() TextViewState {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

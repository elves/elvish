// Package listbox implements a widget for displaying and navigating a list of
// items.
package listbox

import (
	"strings"
	"sync"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Widget supports displaying a list of items, including scrolling and selecting
// functions. It implements the clitypes.Widget interface. An empty Widget is
// directly usable.
type Widget struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	// Publically accessible state.
	State State
	// An Handler that takes precedence over the default handling of events.
	OverlayHandler clitypes.Handler
	// A placeholder to show when there are no items.
	Placeholder styled.Text
	// A function called on the accept event.
	OnAccept func(i int)
}

var _ = clitypes.Widget(&Widget{})

// AddOverlay sets the OverlayHandler of w using the return value of the
// function. It returns w itself and is mainly useful for building a Widget
// instance in a single expression.
func (w *Widget) AddOverlay(overlay func(*Widget) clitypes.Handler) *Widget {
	w.OverlayHandler = overlay(w)
	return w
}

func (w *Widget) init() {
	if w.OverlayHandler == nil {
		w.OverlayHandler = clitypes.DummyHandler{}
	}
	if w.OnAccept == nil {
		w.OnAccept = func(i int) {}
	}
}

var styleForSelected = "inverse"

func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()
	w.StateMutex.Lock()
	s := &w.State
	items, selected, lastFirst := s.Items, s.Selected, s.LastFirst

	if items == nil || items.Len() == 0 {
		s.LastFirst = -1
		w.StateMutex.Unlock()
		return layout.Label{w.Placeholder}.Render(width, height)
	}

	n := items.Len()
	first, firstCrop := findWindow(items, selected, lastFirst, height)
	s.LastFirst = first
	w.StateMutex.Unlock()

	allLines := []styled.Text{}
	hasCropped := firstCrop > 0

	var i int
	for i = first; i < n && len(allLines) < height; i++ {
		item := items.Show(i)
		lines := item.SplitByRune('\n')
		if i == first {
			lines = lines[firstCrop:]
		}
		if i == selected {
			for j := range lines {
				lines[j] = styled.Transform(
					lines[j].ConcatText(styled.Plain(strings.Repeat(" ", width))),
					styleForSelected)
			}
		}
		// TODO: Optionally, add underlines to the last line as a visual
		// separator between adjacent entries.

		if len(allLines)+len(lines) > height {
			lines = lines[:len(allLines)+len(lines)-height]
			hasCropped = true
		}
		allLines = append(allLines, lines...)
	}

	var rd clitypes.Renderer = layout.CroppedLines{allLines}
	if first > 0 || i < n || hasCropped {
		rd = layout.VScrollbarContainer{rd, layout.VScrollbar{n, first, i}}
	}
	return rd.Render(width, height)
}

func (w *Widget) Handle(event term.Event) bool {
	w.init()

	if w.OverlayHandler.Handle(event) {
		return true
	}

	switch event {
	case term.K(ui.Up):
		w.MutateListboxState(func(s *State) {
			switch {
			case s.Selected >= s.Items.Len():
				s.Selected = s.Items.Len() - 1
			case s.Selected <= 0:
				s.Selected = 0
			default:
				s.Selected--
			}
		})
		return true
	case term.K(ui.Down):
		w.MutateListboxState(func(s *State) {
			switch {
			case s.Selected >= s.Items.Len()-1:
				s.Selected = s.Items.Len() - 1
			case s.Selected < 0:
				s.Selected = 0
			default:
				s.Selected++
			}
		})
		return true
	case term.K(ui.Enter):
		w.StateMutex.RLock()
		selected := w.State.Selected
		w.StateMutex.RUnlock()
		w.OnAccept(selected)
		return true
	}
	return false
}

// MutateListboxState calls the given function while locking StateMutex.
func (w *Widget) MutateListboxState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

// CopyListboxState returns a copy of the state.
func (w *Widget) CopyListboxState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

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
	// Publically accessible state.
	State State
	// A placeholder to show when there are no items.
	Placeholder styled.Text
	// A function called on the accept event.
	OnAccept func(i int)
}

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Mutex     sync.RWMutex
	Itemer    Itemer
	NItems    int
	Selected  int
	LastFirst int
}

// Mutate calls the given function while locking the mutex.
func (s *State) Mutate(f func(*State)) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	f(s)
}

// Itemer wraps the Item method.
type Itemer interface {
	// Item returns the item at the given zero-based index.
	Item(i int) styled.Text
}

var _ = clitypes.Widget(&Widget{})

func (w *Widget) init() {
	if w.OnAccept == nil {
		w.OnAccept = func(i int) {}
	}
}

var styleForSelected = "inverse"

func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()
	s := &w.State
	s.Mutex.Lock()
	itemer, n, selected, lastFirst := s.Itemer, s.NItems, s.Selected, s.LastFirst

	if itemer == nil || n == 0 {
		s.LastFirst = -1
		s.Mutex.Unlock()
		return layout.Label{w.Placeholder}.Render(width, height)
	}

	first, firstCrop := findWindow(itemer, n, selected, lastFirst, height)
	s.LastFirst = first
	s.Mutex.Unlock()

	allLines := []styled.Text{}
	hasCropped := firstCrop > 0

	var i int
	for i = first; i < n && len(allLines) < height; i++ {
		item := itemer.Item(i)
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
	switch event {
	case term.K(ui.Up):
		w.State.Mutate(func(s *State) {
			switch {
			case s.Selected >= s.NItems:
				s.Selected = s.NItems - 1
			case s.Selected <= 0:
				s.Selected = 0
			default:
				s.Selected--
			}
		})
	case term.K(ui.Down):
		w.State.Mutate(func(s *State) {
			switch {
			case s.Selected >= s.NItems-1:
				s.Selected = s.NItems - 1
			case s.Selected < 0:
				s.Selected = 0
			default:
				s.Selected++
			}
		})
	case term.K(ui.Enter):
		w.State.Mutex.RLock()
		selected := w.State.Selected
		w.State.Mutex.RUnlock()
		w.OnAccept(selected)
	}
	return false
}

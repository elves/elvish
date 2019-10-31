// Package colview provides a widget for arranging several widgets in a column.
package colview

import (
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Widget is a colview widget.
type Widget interface {
	el.Widget
	// MutateColViewState mutates the state.
	MutateColViewState(f func(*State))
	// CopyColViewState returns a copy of the state.
	CopyColViewState() State
	// Left triggers the OnLeft callback.
	Left()
	// Right triggers the OnRight callback.
	Right()
}

// Spec specifies the configuration and initial state.
type Spec struct {
	// An overlay handler.
	OverlayHandler el.Handler
	// A function that takes the number of columns and return weights for the
	// widths of the columns. The returned slice must have a size of n. If this
	// function is nil, all the columns will have the same weight.
	Weights func(n int) []int
	// A function called when the Left method of Widget is called, or when Left
	// is pressed and unhandled.
	OnLeft func(w Widget)
	// A function called when the Right method of Widget is called, or when
	// Right is pressed and unhandled.
	OnRight func(w Widget)

	// State. Specifies the initial state when used in New.
	State State
}

// State keeps the state of the colview widget.
type State struct {
	Columns     []el.Widget
	FocusColumn int
}

type widget struct {
	// Mutex for synchronizing access to State.
	StateMutex sync.RWMutex
	Spec
}

// New creates a Widget from the given specification.
func New(spec Spec) Widget {
	if spec.OverlayHandler == nil {
		spec.OverlayHandler = el.DummyHandler{}
	}
	if spec.Weights == nil {
		spec.Weights = equalWeights
	}
	if spec.OnLeft == nil {
		spec.OnLeft = func(Widget) {}
	}
	if spec.OnRight == nil {
		spec.OnRight = func(Widget) {}
	}
	return &widget{Spec: spec}
}

func equalWeights(n int) []int {
	weights := make([]int, n)
	for i := 0; i < n; i++ {
		weights[i] = 1
	}
	return weights
}

func (w *widget) MutateColViewState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

func (w *widget) CopyColViewState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

const colGap = 1

// Render renders all the columns side by side, putting the dot in the focused
// column.
func (w *widget) Render(width, height int) *ui.Buffer {
	state := w.CopyColViewState()
	ncols := len(state.Columns)
	if ncols == 0 {
		// No column.
		return &ui.Buffer{Width: width}
	}
	if width < ncols {
		// To narrow; give up by rendering nothing.
		return &ui.Buffer{Width: width}
	}
	colWidths := distribute(width-(ncols-1)*colGap, w.Weights(ncols))
	var buf ui.Buffer
	for i, col := range state.Columns {
		if i > 0 {
			buf.Width += colGap
		}
		bufCol := col.Render(colWidths[i], height)
		buf.ExtendRight(bufCol)
	}
	return &buf
}

// Handle handles the event first by consulting the overlay handler, and then
// delegating the event to the currently focused column.
func (w *widget) Handle(event term.Event) bool {
	if w.OverlayHandler.Handle(event) {
		return true
	}
	state := w.CopyColViewState()
	if 0 <= state.FocusColumn && state.FocusColumn < len(state.Columns) {
		if state.Columns[state.FocusColumn].Handle(event) {
			return true
		}
	}

	switch event {
	case term.K(ui.Left):
		w.Left()
		return true
	case term.K(ui.Right):
		w.Right()
		return true
	default:
		return false
	}
}

func (w *widget) Left() {
	w.OnLeft(w)
}

func (w *widget) Right() {
	w.OnRight(w)
}

// Distributes fullWidth according to the weights, rounding to integers.
//
// This works iteratively each step by taking the sum of all remaining weights,
// and using floor(remainedWidth * currentWeight / remainedAllWeights) for the
// current column.
//
// A simpler alternative is to simply use floor(fullWidth * currentWeight /
// allWeights) at each step, and also giving the remainder to the last column.
// However, this means that the last column gets all the rounding errors from
// flooring, which can be big. The more sophisticated algorithm distributes the
// rounding errors among all the remaining elements and can result in a much
// better distribution, and as a special upside, does not need to handle the
// last column as a special case.
//
// As an extreme example, consider the case of fullWidth = 9, weights = {1, 1,
// 1, 1, 1} (five 1's). Using the simplistic algorithm, the widths are {1, 1, 1,
// 1, 5}. Using the more complex algorithm, the widths are {1, 2, 2, 2, 2}.
func distribute(fullWidth int, weights []int) []int {
	remainedWidth := fullWidth
	remainedWeight := 0
	for _, weight := range weights {
		remainedWeight += weight
	}

	widths := make([]int, len(weights))
	for i, weight := range weights {
		widths[i] = remainedWidth * weight / remainedWeight
		remainedWidth -= widths[i]
		remainedWeight -= weight
	}
	return widths
}

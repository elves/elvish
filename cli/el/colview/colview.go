// Package colview provides a widget for arranging several widgets in a column.
package colview

import (
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Widget is a colview widget.
type Widget struct {
	// Mutex for synchronizing access to State.
	StateMutex sync.RWMutex
	// Public state. Access concurrent to either of Widget's methods must be
	// synchronized using StateMutex.
	State State

	// An overlay handler.
	OverlayHandler el.Handler
	// A function that takes the number of columns and return weights for the
	// widths of the columns. The returned slice must have a size of n. If this
	// function is nil, all the columns will have the same weight.
	Weights func(n int) []int
}

var _ = el.Widget(&Widget{})

// State keeps the state of the colview widget.
type State struct {
	Columns     []el.Widget
	FocusColumn int
}

func equalWeights(n int) []int {
	weights := make([]int, n)
	for i := 0; i < n; i++ {
		weights[i] = 1
	}
	return weights
}

// Initializes nil members to sensible default values. This method is called at
// the beginning of most public methods.
func (w *Widget) init() {
	if w.OverlayHandler == nil {
		w.OverlayHandler = el.DummyHandler{}
	}
	if w.Weights == nil {
		w.Weights = equalWeights
	}
}

// MutateColViewState calls the given function with a pointer to State, while
// locking StateMutex.
func (w *Widget) MutateColViewState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

// CopyColViewState returns a copy of the state.
func (w *Widget) CopyColViewState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

const colGap = 1

// Render renders all the columns side by side, putting the dot in the focused
// column.
func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()

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
func (w *Widget) Handle(event term.Event) bool {
	w.init()

	if w.OverlayHandler.Handle(event) {
		return true
	}
	state := w.CopyColViewState()
	if 0 <= state.FocusColumn && state.FocusColumn < len(state.Columns) {
		return state.Columns[state.FocusColumn].Handle(event)
	}
	return false
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

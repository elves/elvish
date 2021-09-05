package tk

import (
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// ColView is a Widget that arranges several widgets in a column.
type ColView interface {
	Widget
	// MutateState mutates the state.
	MutateState(f func(*ColViewState))
	// CopyState returns a copy of the state.
	CopyState() ColViewState
	// Left triggers the OnLeft callback.
	Left()
	// Right triggers the OnRight callback.
	Right()
}

// ColViewSpec specifies the configuration and initial state for ColView.
type ColViewSpec struct {
	// Key bindings.
	Bindings Bindings
	// A function that takes the number of columns and return weights for the
	// widths of the columns. The returned slice must have a size of n. If this
	// function is nil, all the columns will have the same weight.
	Weights func(n int) []int
	// A function called when the Left method of Widget is called, or when Left
	// is pressed and unhandled.
	OnLeft func(w ColView)
	// A function called when the Right method of Widget is called, or when
	// Right is pressed and unhandled.
	OnRight func(w ColView)

	// State. Specifies the initial state when used in New.
	State ColViewState
}

// ColViewState keeps the mutable state of the ColView widget.
type ColViewState struct {
	Columns     []Widget
	FocusColumn int
}

type colView struct {
	// Mutex for synchronizing access to State.
	StateMutex sync.RWMutex
	ColViewSpec
}

// NewColView creates a new ColView from the given spec.
func NewColView(spec ColViewSpec) ColView {
	if spec.Bindings == nil {
		spec.Bindings = DummyBindings{}
	}
	if spec.Weights == nil {
		spec.Weights = equalWeights
	}
	if spec.OnLeft == nil {
		spec.OnLeft = func(ColView) {}
	}
	if spec.OnRight == nil {
		spec.OnRight = func(ColView) {}
	}
	return &colView{ColViewSpec: spec}
}

func equalWeights(n int) []int {
	weights := make([]int, n)
	for i := 0; i < n; i++ {
		weights[i] = 1
	}
	return weights
}

func (w *colView) MutateState(f func(*ColViewState)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

func (w *colView) CopyState() ColViewState {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	copied := w.State
	copied.Columns = append([]Widget(nil), w.State.Columns...)
	return copied
}

const colViewColGap = 1

// Render renders all the columns side by side, putting the dot in the focused
// column.
func (w *colView) Render(width, height int) *term.Buffer {
	cols, widths := w.prepareRender(width)
	if len(cols) == 0 {
		return &term.Buffer{Width: width}
	}
	var buf term.Buffer
	for i, col := range cols {
		if i > 0 {
			buf.Width += colViewColGap
		}
		bufCol := col.Render(widths[i], height)
		buf.ExtendRight(bufCol)
	}
	return &buf
}

func (w *colView) MaxHeight(width, height int) int {
	cols, widths := w.prepareRender(width)
	max := 0
	for i, col := range cols {
		colMax := col.MaxHeight(widths[i], height)
		if max < colMax {
			max = colMax
		}
	}
	return max
}

// Returns widgets in and widths of columns.
func (w *colView) prepareRender(width int) ([]Widget, []int) {
	state := w.CopyState()
	ncols := len(state.Columns)
	if ncols == 0 {
		// No column.
		return nil, nil
	}
	if width < ncols {
		// To narrow; give up by rendering nothing.
		return nil, nil
	}
	widths := distribute(width-(ncols-1)*colViewColGap, w.Weights(ncols))
	return state.Columns, widths
}

// Handle handles the event first by consulting the overlay handler, and then
// delegating the event to the currently focused column.
func (w *colView) Handle(event term.Event) bool {
	if w.Bindings.Handle(w, event) {
		return true
	}
	state := w.CopyState()
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

func (w *colView) Left() {
	w.OnLeft(w)
}

func (w *colView) Right() {
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

// Package abstract provides an abstract command-line editor and defines
// interfaces for the concrete functionalities it depends on.
package abstract

// Editor is an abstract command-line editor. It implements a UI main loop, and
// relies on callbacks for concrete functionalities -- setting up, reading and
// writing terminal and handling terminal events.
type Editor struct {
	setupCb  SetupCb
	inputCb  InputCb
	redrawCb RedrawCb
	handleCb HandleCb
}

// Event is a placeholder type for terminal events. Should Go support generic
// typing, this will be a type parameter on Editor.
type Event interface{}

// SetupCb sets up the terminal for the editor, and returns a callback for
// undoing the setup and any errors. It should only return an error when the
// terminal is completely unsuitable for subsequent operations. Nonfatal errors
// can be printed directly to the terminal.
type SetupCb func() (undo func(), err error)

// InputCb provides terminal events for the Editor. It should return a channel
// of events, and a function for stopping reading events from the terminal.
type InputCb func() (eventCh <-chan Event, stop func())

// RedrawCb redraws the editor UI to the terminal.
type RedrawCb func(flag RedrawFlag)

// RedrawFlag carries bit flags for RedrawCb.
type RedrawFlag uint

// Bit flags for RedrawFlag.
const (
	// FullRedraw signals a "full redraw". This means that the the terminal is
	// not necessarily left in the same state after the last redraw happened.
	// Hence, if the redrawer only applies a delta to the terminal by default,
	// when this flag is on it should clear the "canvas" and render the UI in
	// full. This flag is on when RedrawCb is called for the first time in a
	// Read loop.
	FullRedraw RedrawFlag = 1 << iota
	// FinishRedraw signals that this is the finishing redraw in a Read loop.
	// This means that what is drawn onto the terminal will remain in the
	// loopback buffer and cannot be modified anymore. The redrawer may only
	// want to draw the essential UI parts, like prompt and buffer.
	FinishRedraw
)

// HandleCb handles a terminal event. If quit is true, Read returns with buffer.
type HandleCb func(event Event) (buffer string, quit bool)

func NewEditor(inputCb InputCb, handleCb HandleCb) *Editor {
	return &Editor{
		inputCb:  inputCb,
		handleCb: handleCb,
	}
}

func (ed *Editor) SetupCb(cb SetupCb) {
}

func (ed *Editor) RedrawCb(cb RedrawCb) {
}

func (ed *Editor) Redraw(full bool) {
}

// Read reads and processes terminal events, until HandleCb requests it to quit.
// It only manages the event loop, and delegates concrete work to callbacks. It
// is fully serial: it does not spawn any goroutines and never calls two
// callbacks in parallel, so the callbacks may manipulate shared states without
// synchronization.
func (ed *Editor) Read() (buffer string, err error) {
	return "", nil
}

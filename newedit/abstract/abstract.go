// Package abstract provides an abstract command-line editor and defines
// interfaces for the concrete functionalities it depends on.
package abstract

import "sync"

const inputChBuffer = 10

// Editor is an abstract command-line editor. It implements a UI main loop, and
// relies on callbacks for concrete functionalities -- setting up, reading and
// writing terminal and handling terminal events.
type Editor struct {
	inputCh  chan Event
	handleCb HandleCb

	setupCb  SetupCb
	redrawCb RedrawCb

	redrawCh    chan struct{}
	redrawFull  bool
	redrawMutex *sync.Mutex
}

// Event is a placeholder type for terminal events. Should Go support generic
// typing, this will be a type parameter on Editor.
type Event interface{}

// SetupCb sets up the terminal for the editor, and returns a callback for
// undoing the setup and any errors. It should only return an error when the
// terminal is completely unsuitable for subsequent operations. Nonfatal errors
// can be printed directly to the terminal.
type SetupCb func() (undo func(), err error)

func dummySetupCb() (func(), error) { return nil, nil }

// RedrawCb redraws the editor UI to the terminal.
type RedrawCb func(flag RedrawFlag)

func dummyRedrawCb(RedrawFlag) {}

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
	// FinalRedraw signals that this is the finishing redraw in a Read loop.
	// This means that what is drawn onto the terminal will remain in the
	// loopback buffer and cannot be modified anymore. The redrawer may only
	// want to draw the essential UI parts, like prompt and buffer.
	FinalRedraw
)

// HandleCb handles a terminal event. If quit is true, Read returns with buffer.
type HandleCb func(event Event) (buffer string, quit bool)

// NewEditor creates a new Editor.
func NewEditor(handleCb HandleCb) *Editor {
	return &Editor{
		inputCh:  make(chan Event, inputChBuffer),
		handleCb: handleCb,

		setupCb:  dummySetupCb,
		redrawCb: dummyRedrawCb,

		redrawCh:    make(chan struct{}, 1),
		redrawFull:  false,
		redrawMutex: new(sync.Mutex),
	}
}

// SetupCb sets the setup callback. It must be called before any Read call.
func (ed *Editor) SetupCb(cb SetupCb) {
	ed.setupCb = cb
}

// RedrawCb sets the redraw callback. It must be called before any Read call.
func (ed *Editor) RedrawCb(cb RedrawCb) {
	ed.redrawCb = cb
}

// Redraw requests a redraw. If full is true, a full redraw is requested. It
// never blocks.
func (ed *Editor) Redraw(full bool) {
	ed.redrawMutex.Lock()
	defer ed.redrawMutex.Unlock()
	if full {
		ed.redrawFull = true
	}
	select {
	case ed.redrawCh <- struct{}{}:
	default:
	}
}

// Input provides an input event.
func (ed *Editor) Input(event Event) {
	ed.inputCh <- event
}

// Read reads and processes terminal events, until HandleCb requests it to quit.
// It only manages the event loop, and delegates concrete work to callbacks. It
// is fully serial: it does not spawn any goroutines and never calls two
// callbacks in parallel, so the callbacks may manipulate shared states without
// synchronization.
func (ed *Editor) Read() (buffer string, err error) {
	restore, err := ed.setupCb()
	if err != nil {
		return "", err
	}
	if restore != nil {
		defer restore()
	}

	for {
		var redrawFlag RedrawFlag
		if ed.extractRedrawFull() {
			redrawFlag |= FullRedraw
		}
		ed.redrawCb(redrawFlag)
		select {
		case event := <-ed.inputCh:
			buffer, quit := ed.handleCb(event)
			if quit {
				ed.redrawCb(FinalRedraw)
				return buffer, nil
			}
		case <-ed.redrawCh:
		}
	}
}

func (ed *Editor) extractRedrawFull() bool {
	ed.redrawMutex.Lock()
	defer ed.redrawMutex.Unlock()

	full := ed.redrawFull
	ed.redrawFull = false
	return full
}

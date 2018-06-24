// Package loop provides a generic main loop for an editor.
package loop

import "sync"

// Buffer size of the input channel. The value is chosen for no particular
// reason.
const inputChSize = 128

// Loop implements a generic main loop for an editor.
type Loop struct {
	inputCh  chan Event
	handleCb HandleCb

	redrawCb RedrawCb

	redrawCh    chan struct{}
	redrawFull  bool
	redrawMutex *sync.Mutex
}

// Event is a placeholder type for terminal events. Should Go support generic
// typing, this will be a type parameter on EditLoop.
type Event interface{}

// RedrawCb is a callback for redrawing the editor UI to the terminal.
type RedrawCb func(flag RedrawFlag)

func dummyRedrawCb(RedrawFlag) {}

// RedrawFlag carries bit flags for RedrawCb.
type RedrawFlag uint

// Bit flags for RedrawFlag.
const (
	// FullRedraw signals a "full redraw". This is set on the first RedrawCb
	// call or when Redraw has been called with full = true.
	FullRedraw RedrawFlag = 1 << iota
	// FinalRedraw signals that this is the final redraw in the event loop.
	FinalRedraw
)

// HandleCb is a callback for handling a terminal event. If quit is true, Read
// returns with buffer.
type HandleCb func(event Event) (buffer string, quit bool)

func dummyHandleCb(Event) (string, bool) { return "", false }

// New creates a new Loop instance.
func New() *Loop {
	return &Loop{
		inputCh:  make(chan Event, inputChSize),
		handleCb: dummyHandleCb,
		redrawCb: dummyRedrawCb,

		redrawCh:    make(chan struct{}, 1),
		redrawFull:  false,
		redrawMutex: new(sync.Mutex),
	}
}

// HandleCb sets the handle callback. It must be called before any Read call.
func (ed *Loop) HandleCb(cb HandleCb) {
	ed.handleCb = cb
}

// RedrawCb sets the redraw callback. It must be called before any Read call.
func (ed *Loop) RedrawCb(cb RedrawCb) {
	ed.redrawCb = cb
}

// Redraw requests a redraw. If full is true, a full redraw is requested. It
// never blocks.
func (ed *Loop) Redraw(full bool) {
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

// Input provides an input event. It may block if the internal event buffer is
// full.
func (ed *Loop) Input(event Event) {
	ed.inputCh <- event
}

// Run runs the event loop, until an event causes HandleCb to return quit =
// true. It is generic and delegates all concrete work to callbacks. It is fully
// serial: it does not spawn any goroutines and never calls two callbacks in
// parallel, so the callbacks may manipulate shared states without
// synchronization.
func (ed *Loop) Run() (buffer string, err error) {
	for {
		var redrawFlag RedrawFlag
		if ed.extractRedrawFull() {
			redrawFlag |= FullRedraw
		}
		ed.redrawCb(redrawFlag)
		select {
		case event := <-ed.inputCh:
			// Consume all events in the channel to minimize redraws.
		consumeAllEvents:
			for {
				buffer, quit := ed.handleCb(event)
				if quit {
					ed.redrawCb(FinalRedraw)
					return buffer, nil
				}
				select {
				case event = <-ed.inputCh:
					// Continue the loop of consuming all events.
				default:
					break consumeAllEvents
				}
			}
		case <-ed.redrawCh:
		}
	}
}

func (ed *Loop) extractRedrawFull() bool {
	ed.redrawMutex.Lock()
	defer ed.redrawMutex.Unlock()

	full := ed.redrawFull
	ed.redrawFull = false
	return full
}

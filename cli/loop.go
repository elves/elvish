package cli

import "sync"

// Buffer size of the input channel. The value is chosen for no particular
// reason.
const inputChSize = 128

// A generic main loop manager.
type loop struct {
	inputCh  chan event
	handleCb handleCb

	redrawCb redrawCb

	redrawCh    chan struct{}
	redrawFull  bool
	redrawMutex *sync.Mutex

	returnCh chan loopReturn
}

type loopReturn struct {
	buffer string
	err    error
}

// A placeholder type for events.
type event interface{}

// Callback for redrawing the editor UI to the terminal.
type redrawCb func(flag redrawFlag)

func dummyRedrawCb(redrawFlag) {}

// Flag to redrawCb.
type redrawFlag uint

// Bit flags for redrawFlag.
const (
	// fullRedraw signals a "full redraw". This is set on the first RedrawCb
	// call or when Redraw has been called with full = true.
	fullRedraw redrawFlag = 1 << iota
	// finalRedraw signals that this is the final redraw in the event loop.
	finalRedraw
)

// Callback for handling a terminal event.
type handleCb func(event)

func dummyHandleCb(event) {}

// newLoop creates a new Loop instance.
func newLoop() *loop {
	return &loop{
		inputCh:  make(chan event, inputChSize),
		handleCb: dummyHandleCb,
		redrawCb: dummyRedrawCb,

		redrawCh:    make(chan struct{}, 1),
		redrawFull:  false,
		redrawMutex: new(sync.Mutex),

		returnCh: make(chan loopReturn, 1),
	}
}

// HandleCb sets the handle callback. It must be called before any Read call.
func (ed *loop) HandleCb(cb handleCb) {
	ed.handleCb = cb
}

// RedrawCb sets the redraw callback. It must be called before any Read call.
func (ed *loop) RedrawCb(cb redrawCb) {
	ed.redrawCb = cb
}

// Redraw requests a redraw. If full is true, a full redraw is requested. It
// never blocks.
func (ed *loop) Redraw(full bool) {
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
func (ed *loop) Input(ev event) {
	ed.inputCh <- ev
}

// Return requests the main loop to return. It never blocks. If Return has been
// called before during the current loop iteration, it has no effect.
func (ed *loop) Return(buffer string, err error) {
	select {
	case ed.returnCh <- loopReturn{buffer, err}:
	default:
	}
}

// HasReturned returns whether Return has been called during the current loop
// iteration.
func (ed *loop) HasReturned() bool {
	return len(ed.returnCh) == 1
}

// Run runs the event loop, until the Return method is called. It is generic
// and delegates all concrete work to callbacks. It is fully serial: it does
// not spawn any goroutines and never calls two callbacks in parallel, so the
// callbacks may manipulate shared states without synchronization.
func (ed *loop) Run() (buffer string, err error) {
	for {
		var flag redrawFlag
		if ed.extractRedrawFull() {
			flag |= fullRedraw
		}
		ed.redrawCb(flag)
		select {
		case event := <-ed.inputCh:
			// Consume all events in the channel to minimize redraws.
		consumeAllEvents:
			for {
				ed.handleCb(event)
				select {
				case ret := <-ed.returnCh:
					ed.redrawCb(finalRedraw)
					return ret.buffer, ret.err
				default:
				}
				select {
				case event = <-ed.inputCh:
					// Continue the loop of consuming all events.
				default:
					break consumeAllEvents
				}
			}
		case ret := <-ed.returnCh:
			ed.redrawCb(finalRedraw)
			return ret.buffer, ret.err
		case <-ed.redrawCh:
		}
	}
}

func (ed *loop) extractRedrawFull() bool {
	ed.redrawMutex.Lock()
	defer ed.redrawMutex.Unlock()

	full := ed.redrawFull
	ed.redrawFull = false
	return full
}

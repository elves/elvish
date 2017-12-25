package tty

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/sys"
	"golang.org/x/sys/windows"
)

// XXX Put here to make edit package build on Windows.
const DefaultSeqTimeout = 10 * time.Millisecond

type reader struct {
	console   windows.Handle
	eventChan chan Event

	stopEvent   windows.Handle
	stopChan    chan struct{}
	stopAckChan chan struct{}
}

// NewReader creates a new Reader instance.
func newReader(file *os.File) Reader {
	console, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		panic(fmt.Errorf("GetStdHandle(STD_INPUT_HANDLE): %v", err))
	}
	stopEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		panic(fmt.Errorf("CreateEvent: %v", err))
	}
	return &reader{
		console, make(chan Event), stopEvent, nil, nil}
}

func (r *reader) SetRaw(bool) {
	// NOP on Windows.
}

func (r *reader) EventChan() <-chan Event {
	return r.eventChan
}

func (r *reader) Start() {
	r.stopChan = make(chan struct{})
	r.stopAckChan = make(chan struct{})
	go r.run()
}

var errNr0 = errors.New("ReadConsoleInput reads 0 input")

func (r *reader) run() {
	handles := []windows.Handle{r.console, r.stopEvent}
	for {
		triggered, _, err := sys.WaitForMultipleObjects(handles, false, sys.INFINITE)
		if err != nil {
			r.fatal(err)
			return
		}
		if triggered == 1 {
			<-r.stopChan
			close(r.stopAckChan)
			return
		}

		var buf [1]sys.InputRecord
		nr, err := sys.ReadConsoleInput(r.console, buf[:])
		if nr == 0 {
			r.fatal(errNr0)
			return
		}
		if err != nil {
			r.fatal(err)
			return
		}
		event := convertEvent(buf[0].GetEvent())
		if event != nil {
			r.send(event)
		}
	}
}

func (r *reader) nonFatal(err error) {
	r.send(NonfatalErrorEvent{err})
}

func (r *reader) fatal(err error) {
	if !r.send(FatalErrorEvent{err}) {
		<-r.stopChan
		close(r.stopAckChan)
		r.resetStopEvent()
	}
}

func (r *reader) send(event Event) (stopped bool) {
	select {
	case r.eventChan <- event:
		return false
	case <-r.stopChan:
		close(r.stopAckChan)
		r.resetStopEvent()
		return true
	}
}

func (r *reader) resetStopEvent() {
	err := windows.ResetEvent(r.stopEvent)
	if err != nil {
		panic(err)
	}
}

func (r *reader) Stop() {
	err := windows.SetEvent(r.stopEvent)
	if err != nil {
		log.Println("SetEvent:", err)
	}
	close(r.stopChan)
	<-r.stopAckChan
}

func (r *reader) Close() {
	err := windows.CloseHandle(r.stopEvent)
	if err != nil {
		log.Println("Closing stopEvent handle for reader:", err)
	}
	close(r.eventChan)
}

// A subset of virtual key codes listed in
// https://msdn.microsoft.com/en-us/library/windows/desktop/dd375731(v=vs.85).aspx
var keyCodeToRune = map[uint16]rune{
	0x08: ui.Backspace, 0x09: ui.Tab,
	0x0d: ui.Enter,
	0x1b: '\x1b',
	0x20: ' ',
	0x23: ui.End, 0x24: ui.Home,
	0x25: ui.Left, 0x26: ui.Up, 0x27: ui.Right, 0x28: ui.Down,
	0x2d: ui.Insert, 0x2e: ui.Delete,
	/* 0x30 - 0x39: digits, same with ASCII */
	/* 0x41 - 0x5a: letters, same with ASCII */
	/* 0x60 - 0x6f: numpads; currently ignored */
	0x70: ui.F1, 0x71: ui.F2, 0x72: ui.F3, 0x73: ui.F4, 0x74: ui.F5, 0x75: ui.F6,
	0x76: ui.F7, 0x77: ui.F8, 0x78: ui.F9, 0x79: ui.F10, 0x7a: ui.F11, 0x7b: ui.F12,
	/* 0x7c - 0x87: F13 - F24; currently ignored */
	0xba: ';', 0xbb: '=', 0xbc: ',', 0xbd: '-', 0xbe: '.', 0xbf: '/', 0xc0: '`',
	0xdb: '[', 0xdc: '\\', 0xdd: ']', 0xde: '\'',
}

// A subset of constants listed in
// https://docs.microsoft.com/en-us/windows/console/key-event-record-str
const (
	leftAlt   = 0x02
	leftCtrl  = 0x08
	rightAlt  = 0x01
	rightCtrl = 0x04
	shift     = 0x10
)

// convertEvent converts the native sys.InputEvent type to a suitable Event
// type. It returns nil if the event should be ignored.
func convertEvent(event sys.InputEvent) Event {
	switch event := event.(type) {
	case *sys.KeyEvent:
		if event.BKeyDown == 0 {
			// Ignore keyup events.
			return nil
		}
		r := rune(event.UChar[0]) + rune(event.UChar[1])<<8
		if event.DwControlKeyState == 0 {
			// No modifier
			// TODO: Deal with surrogate pairs
			if 0x20 <= r && r != 0x7f {
				return KeyEvent(ui.Key{Rune: r})
			}
		} else if event.DwControlKeyState == shift {
			// If only the shift is held down, we try and see if this is a
			// non-functional key by looking if the rune generated is a
			// printable ASCII character.
			if 0x20 <= r && r < 0x7f {
				return KeyEvent(ui.Key{Rune: r})
			}
		}
		r = convertRune(event.WVirtualKeyCode)
		if r == 0 {
			return nil
		}
		mod := convertMod(event.DwControlKeyState)
		return KeyEvent(ui.Key{Rune: r, Mod: mod})
	//case *sys.MouseEvent:
	//case *sys.WindowBufferSizeEvent:
	default:
		// Other events are ignored.
		return nil
	}
}

func convertRune(keyCode uint16) rune {
	r, ok := keyCodeToRune[keyCode]
	if ok {
		return r
	}
	if '0' <= keyCode && keyCode <= '9' || 'A' <= keyCode && keyCode <= 'Z' {
		return rune(keyCode)
	}
	return 0
}

func convertMod(state uint32) ui.Mod {
	mod := ui.Mod(0)
	if state&(leftAlt|rightAlt) != 0 {
		mod |= ui.Alt
	}
	if state&(leftCtrl|rightCtrl) != 0 {
		mod |= ui.Ctrl
	}
	if state&shift != 0 {
		mod |= ui.Shift
	}
	return mod
}

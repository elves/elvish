package term

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/elves/elvish/pkg/sys"
	"github.com/elves/elvish/pkg/ui"
	"golang.org/x/sys/windows"
)

// TODO(xiaq): Put here to make edit package build on Windows. Refactor so
// that isn't needed?
const DefaultSeqTimeout = 10 * time.Millisecond

type reader struct {
	console   windows.Handle
	stopEvent windows.Handle
	// A mutex that is held during ReadEvent.
	mutex sync.Mutex
}

// Creates a new Reader instance.
func newReader(file *os.File) Reader {
	console, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		panic(fmt.Errorf("GetStdHandle(STD_INPUT_HANDLE): %v", err))
	}
	stopEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		panic(fmt.Errorf("CreateEvent: %v", err))
	}
	return &reader{console: console, stopEvent: stopEvent}
}

func (r *reader) ReadEvent() (Event, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	handles := []windows.Handle{r.console, r.stopEvent}
	for {
		triggered, _, err := sys.WaitForMultipleObjects(handles, false, sys.INFINITE)
		if err != nil {
			return nil, err
		}
		if triggered == 1 {
			return nil, ErrStopped
		}

		var buf [1]sys.InputRecord
		nr, err := sys.ReadConsoleInput(r.console, buf[:])
		if nr == 0 {
			return nil, io.ErrNoProgress
		}
		if err != nil {
			return nil, err
		}
		event := convertEvent(buf[0].GetEvent())
		if event != nil {
			return event, nil
		}
		// Got an event that should be ignored; keep going.
	}
}

func (r *reader) ReadRawEvent() (Event, error) {
	return r.ReadEvent()
}

func (r *reader) Close() {
	err := windows.SetEvent(r.stopEvent)
	if err != nil {
		log.Println("SetEvent:", err)
	}
	r.mutex.Lock()
	r.mutex.Unlock()
	err = windows.CloseHandle(r.stopEvent)
	if err != nil {
		log.Println("Closing stopEvent handle for reader:", err)
	}
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
		filteredMod := event.DwControlKeyState & (leftAlt | leftCtrl | rightAlt | rightCtrl | shift)
		if filteredMod == 0 {
			// No modifier
			// TODO: Deal with surrogate pairs
			if 0x20 <= r && r != 0x7f {
				return KeyEvent(ui.Key{Rune: r})
			}
		} else if filteredMod == shift {
			// If only the shift is held down, we try and see if this is a
			// non-functional key by looking if the rune generated is a
			// printable ASCII character.
			if 0x20 <= r && r < 0x7f {
				return KeyEvent(ui.Key{Rune: r})
			}
		}
		mod := convertMod(filteredMod)
		if mod == 0 && event.WVirtualKeyCode == 0x1b {
			// Special case: Normalize 0x1b to Ctrl-[.
			//
			// TODO(xiaq): This is Unix-centric. Maybe the normalized form
			// should be Escape.
			return KeyEvent(ui.Key{Rune: '[', Mod: ui.Ctrl})
		}
		r = convertRune(event.WVirtualKeyCode, mod)
		if r == 0 {
			return nil
		}
		return KeyEvent(ui.Key{Rune: r, Mod: mod})
	//case *sys.MouseEvent:
	//case *sys.WindowBufferSizeEvent:
	default:
		// Other events are ignored.
		return nil
	}
}

func convertRune(keyCode uint16, mod ui.Mod) rune {
	r, ok := keyCodeToRune[keyCode]
	if ok {
		return r
	}
	if '0' <= keyCode && keyCode <= '9' {
		return rune(keyCode)
	}
	if 'A' <= keyCode && keyCode <= 'Z' {
		// If Ctrl is involved, emulate UNIX's convention and use upper case;
		// otherwise use lower case.
		//
		// TODO(xiaq): This is quite Unix-centric. Maybe we should make the
		// base rune case-insensitive when there are modifiers involved.
		if mod&ui.Ctrl != 0 {
			return rune(keyCode)
		}
		return rune(keyCode - 'A' + 'a')
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

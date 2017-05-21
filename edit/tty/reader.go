package tty

import (
	"fmt"
	"os"
	"time"

	"github.com/elves/elvish/edit/ui"
)

var (
	// EscSequenceTimeout is the amount of time within which runes that make up
	// an escape sequence are supposed to follow each other. Modern terminal
	// emulators send escape sequences very fast, so 10ms is more than
	// sufficient. SSH connections on a slow link might be problematic though.
	EscSequenceTimeout = 10 * time.Millisecond
)

// Special rune values used in the return value of (*Reader).ReadRune.
const (
	// No rune received before specified time.
	runeTimeout rune = -1 - iota
	// Error occurred in AsyncReader. The error is left at the readError field.
	runeReadError
)

// Reader converts a stream of events on separate channels.
type Reader struct {
	ar  *AsyncReader
	raw bool

	unitChan chan ReadUnit
	/*
		rawRuneChan chan rune
		keyChan     chan ui.Key
		cprChan     chan Pos
		mouseChan   chan MouseEvent
		pasteChan   chan bool
	*/
	errChan chan error
	quit    chan struct{}
}

type MouseEvent struct {
	Pos
	Down bool
	// Number of the Button, 0-based. -1 for unknown.
	Button int
	Mod    ui.Mod
}

// NewReader creates a new Reader on the given terminal file.
func NewReader(f *os.File) *Reader {
	rd := &Reader{
		NewAsyncReader(f),
		false,
		make(chan ReadUnit),
		/*
			make(chan rune),
			make(chan ui.Key),
			make(chan Pos),
			make(chan MouseEvent),
			make(chan bool),
		*/
		make(chan error),
		nil,
	}
	return rd
}

// SetRaw turns the raw option on or off. If the reader is in the middle of
// reading one event, it takes effect after this event is fully read.
func (rd *Reader) SetRaw(raw bool) {
	rd.raw = raw
}

// UnitChan returns the channel onto which the Reader writes what it has read.
func (rd *Reader) UnitChan() <-chan ReadUnit {
	return rd.unitChan
}

// ErrorChan returns the channel onto which the Reader writes errors it came
// across during the reading process.
func (rd *Reader) ErrorChan() <-chan error {
	return rd.errChan
}

// Run runs the Reader. It blocks until Quit is called and should be called in
// a separate goroutine.
func (rd *Reader) Run() {
	runes := rd.ar.Chan()
	quit := make(chan struct{})
	rd.quit = quit
	go rd.ar.Run()

	for {
		select {
		case r := <-runes:
			if rd.raw {
				rd.unitChan <- RawRune(r)
			} else {
				rd.readOne(r)
			}
		case <-quit:
			return
		}
	}
}

// Quit terminates the loop of Run.
func (rd *Reader) Quit() {
	rd.ar.Quit()
	close(rd.quit)
}

// Close releases files associated with the Reader. It does not close the file
// used to create it.
func (rd *Reader) Close() {
	rd.ar.Close()
}

// readOne attempts to read one key or CPR, led by a rune already read.
func (rd *Reader) readOne(r rune) {
	var unit ReadUnit
	var err error
	currentSeq := string(r)

	badSeq := func(msg string) {
		err = fmt.Errorf("%s: %q", msg, currentSeq)
	}

	// readRune attempts to read a rune within EscSequenceTimeout. It writes to
	// the err and currentSeq variable in the outer scope.
	readRune :=
		func() rune {
			select {
			case r := <-rd.ar.Chan():
				currentSeq += string(r)
				return r
			case err = <-rd.ar.ErrorChan():
				return runeReadError
			case <-time.After(EscSequenceTimeout):
				return runeTimeout
			}
		}

	defer func() {
		if unit != nil {
			select {
			case rd.unitChan <- unit:
			case <-rd.quit:
			}
		}
		if err != nil {
			select {
			case rd.errChan <- err:
			case <-rd.quit:
			}
		}
	}()

	switch r {
	case 0x1b: // ^[ Escape
		r2 := readRune()
		if r2 == runeTimeout || r2 == runeReadError {
			// Nothing follows. Taken as a lone Escape.
			unit = Key{'[', ui.Ctrl}
			break
		}
		switch r2 {
		case '[':
			// A '[' follows. CSI style function key sequence.
			r = readRune()
			if r == runeTimeout || r == runeReadError {
				unit = Key{'[', ui.Alt}
				return
			}

			nums := make([]int, 0, 2)
			var starter rune

			// Read an optional starter.
			switch r {
			case '<':
				starter = r
				r = readRune()
			case 'M':
				// Mouse event.
				cb := readRune()
				if cb == runeTimeout || cb == runeReadError {
					badSeq("Incomplete mouse event")
					return
				}
				cx := readRune()
				if cx == runeTimeout || cx == runeReadError {
					badSeq("Incomplete mouse event")
					return
				}
				cy := readRune()
				if cy == runeTimeout || cy == runeReadError {
					badSeq("Incomplete mouse event")
					return
				}
				down := true
				button := int(cb & 3)
				if button == 3 {
					down = false
					button = -1
				}
				mod := mouseModify(int(cb))
				unit = MouseEvent{
					Pos{int(cy) - 32, int(cx) - 32}, down, button, mod}
				return
			}
		CSISeq:
			for {
				switch {
				case r == ';':
					nums = append(nums, 0)
				case '0' <= r && r <= '9':
					if len(nums) == 0 {
						nums = append(nums, 0)
					}
					cur := len(nums) - 1
					nums[cur] = nums[cur]*10 + int(r-'0')
				case r == runeTimeout:
					// Incomplete CSI.
					badSeq("Incomplete CSI")
					return
				case r == runeReadError:
					// TODO Also complain about incomplte CSI.
					return
				default: // Treat as a terminator.
					break CSISeq
				}

				r = readRune()
			}
			if starter == 0 && r == 'R' {
				// Cursor position report.
				if len(nums) != 2 {
					badSeq("bad CPR")
					return
				}
				unit = CursorPosition{nums[0], nums[1]}
			} else if starter == '<' && (r == 'm' || r == 'M') {
				// SGR-style mouse event.
				if len(nums) != 3 {
					badSeq("bad SGR mouse event")
					return
				}
				down := r == 'M'
				button := nums[0] & 3
				mod := mouseModify(nums[0])
				unit = MouseEvent{Pos{nums[2], nums[1]}, down, button, mod}
			} else if r == '~' && len(nums) == 1 && (nums[0] == 200 || nums[0] == 201) {
				b := nums[0] == 200
				unit = PasteSetting(b)
			} else {
				k := parseCSI(nums, r, currentSeq)
				if k == (ui.Key{}) {
					badSeq("bad CSI")
				} else {
					unit = Key(k)
				}
			}
		case 'O':
			// An 'O' follows. G3 style function key sequence: read one rune.
			r = readRune()
			if r == runeTimeout || r == runeReadError {
				// Nothing follows after 'O'. Taken as ui.Alt-o.
				unit = Key{'o', ui.Alt}
				return
			}
			r, ok := g3Seq[r]
			if ok {
				unit = Key{r, 0}
			} else {
				badSeq("bad G3")
			}
		default:
			// Something other than '[' or 'O' follows. Taken as an
			// ui.Alt-modified key, possibly also modified by ui.Ctrl.
			k := ctrlModify(r2)
			k.Mod |= ui.Alt
			unit = Key(k)
		}
	default:
		k := ctrlModify(r)
		unit = Key(k)
	}
}

// ctrlModify determines whether a rune corresponds to a ui.Ctrl-modified key and
// returns the ui.Key the rune represents.
func ctrlModify(r rune) ui.Key {
	switch r {
	case 0x0:
		return ui.Key{'`', ui.Ctrl} // ^@
	case 0x1e:
		return ui.Key{'6', ui.Ctrl} // ^^
	case 0x1f:
		return ui.Key{'/', ui.Ctrl} // ^_
	case ui.Tab, ui.Enter, ui.Backspace: // ^I ^J ^?
		return ui.Key{r, 0}
	default:
		// Regular ui.Ctrl sequences.
		if 0x1 <= r && r <= 0x1d {
			return ui.Key{r + 0x40, ui.Ctrl}
		}
	}
	return ui.Key{r, 0}
}

// G3-style key sequences: \eO followed by exactly one character. For instance,
// \eOP is ui.F1.
var g3Seq = map[rune]rune{
	'A': ui.Up, 'B': ui.Down, 'C': ui.Right, 'D': ui.Left,

	// ui.F1-ui.F4: xterm, libvte and tmux
	'P': ui.F1, 'Q': ui.F2,
	'R': ui.F3, 'S': ui.F4,

	// ui.Home and ui.End: libvte
	'H': ui.Home, 'F': ui.End,
}

// Tables for CSI-style key sequences, which are \e[ followed by a list of
// semicolon-delimited numeric arguments, before being concluded by a
// non-numeric, non-semicolon rune.

// CSI-style key sequences that can be identified based on the ending rune. For
// instance, \e[A is ui.Up.
var keyByLast = map[rune]ui.Key{
	'A': {ui.Up, 0}, 'B': {ui.Down, 0},
	'C': {ui.Right, 0}, 'D': {ui.Left, 0},
	'H': {ui.Home, 0}, 'F': {ui.End, 0},
	'Z': {ui.Tab, ui.Shift},
}

// CSI-style key sequences ending with '~' and can be identified based on the
// only number argument. For instance, \e[~ is ui.Home. When they are
// modified, they take two arguments, first being 1 and second identifying the
// modifier (see xtermModify). For instance, \e[1;4~ is Shift-Alt-Home.
var keyByNum0 = map[int]rune{
	1: ui.Home, 2: ui.Insert, 3: ui.Delete, 4: ui.End,
	5: ui.PageUp, 6: ui.PageDown,
	11: ui.F1, 12: ui.F2, 13: ui.F3, 14: ui.F4,
	15: ui.F5, 17: ui.F6, 18: ui.F7, 19: ui.F8,
	20: ui.F9, 21: ui.F10, 23: ui.F11, 24: ui.F12,
}

// CSI-style key sequences ending with '~', with 27 as the first numeric
// argument. For instance, \e[27;9~ is ui.Tab.
//
// The list is taken blindly from tmux source xterm-keys.c. I don't have a
// keyboard-terminal combination that generate such sequences, but assumably
// some PC keyboard with a numpad can.
var keyByNum2 = map[int]rune{
	9: '\t', 13: '\r',
	33: '!', 35: '#', 39: '\'', 40: '(', 41: ')', 43: '+', 44: ',', 45: '-',
	46: '.',
	48: '0', 49: '1', 50: '2', 51: '3', 52: '4', 53: '5', 54: '6', 55: '7',
	56: '8', 57: '9',
	58: ':', 59: ';', 60: '<', 61: '=', 62: '>', 63: ';',
}

// parseCSI parses a CSI-style key sequence.
func parseCSI(nums []int, last rune, seq string) ui.Key {
	if k, ok := keyByLast[last]; ok {
		if len(nums) == 0 {
			// Unmodified: \e[A (ui.Up)
			return k
		} else if len(nums) == 2 && nums[0] == 1 {
			// Modified: \e[1;5A (ui.Ctrl-ui.Up)
			return xtermModify(k, nums[1], seq)
		} else {
			return ui.Key{}
		}
	}

	if last == '~' {
		if len(nums) == 1 || len(nums) == 2 {
			if r, ok := keyByNum0[nums[0]]; ok {
				k := ui.Key{r, 0}
				if len(nums) == 1 {
					// Unmodified: \e[5~ (ui.PageUp)
					return k
				}
				// Modified: \e[5;5~ (ui.Ctrl-ui.PageUp)
				return xtermModify(k, nums[1], seq)
			}
		} else if len(nums) == 3 && nums[0] == 27 {
			if r, ok := keyByNum2[nums[2]]; ok {
				k := ui.Key{r, 0}
				return xtermModify(k, nums[1], seq)
			}
		}
	}

	return ui.Key{}
}

func xtermModify(k ui.Key, mod int, seq string) ui.Key {
	switch mod {
	case 0:
		// do nothing
	case 2:
		k.Mod |= ui.Shift
	case 3:
		k.Mod |= ui.Alt
	case 4:
		k.Mod |= ui.Shift | ui.Alt
	case 5:
		k.Mod |= ui.Ctrl
	case 6:
		k.Mod |= ui.Shift | ui.Ctrl
	case 7:
		k.Mod |= ui.Alt | ui.Ctrl
	case 8:
		k.Mod |= ui.Shift | ui.Alt | ui.Ctrl
	default:
		return ui.Key{}
	}
	return k
}

func mouseModify(n int) ui.Mod {
	var mod ui.Mod
	if n&4 != 0 {
		mod |= ui.Shift
	}
	if n&8 != 0 {
		mod |= ui.Alt
	}
	if n&16 != 0 {
		mod |= ui.Ctrl
	}
	return mod
}

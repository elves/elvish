package edit

import (
	"fmt"
	"os"
	"time"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/uitypes"
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
	// Error occured in AsyncReader. The error is left at the readError field.
	runeReadError
)

// Reader converts a stream of events on separate channels.
type Reader struct {
	ar        *tty.AsyncReader
	keyChan   chan uitypes.Key
	cprChan   chan pos
	mouseChan chan mouseEvent
	errChan   chan error
	pasteChan chan bool
	quit      chan struct{}
}

type mouseEvent struct {
	pos
	down bool
	// Number of the button, 0-based. -1 for unknown.
	button int
	mod    uitypes.Mod
}

// NewReader creates a new Reader on the given terminal file.
func NewReader(f *os.File) *Reader {
	rd := &Reader{
		tty.NewAsyncReader(f),
		make(chan uitypes.Key),
		make(chan pos),
		make(chan mouseEvent),
		make(chan error),
		make(chan bool),
		nil,
	}
	return rd
}

// KeyChan returns the channel onto which the Reader writes Keys it has read.
func (rd *Reader) KeyChan() <-chan uitypes.Key {
	return rd.keyChan
}

// CPRChan returns the channel onto which the Reader writes CPRs it has read.
func (rd *Reader) CPRChan() <-chan pos {
	return rd.cprChan
}

// MouseChan returns the channel onto which the Reader writes mouse events it
// has read.
func (rd *Reader) MouseChan() <-chan mouseEvent {
	return rd.mouseChan
}

func (rd *Reader) PasteChan() <-chan bool {
	return rd.pasteChan
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
			rd.readOne(r)
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
	var k uitypes.Key
	var cpr pos
	var mouse mouseEvent
	var err error
	var paste *bool
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
		if k != (uitypes.Key{}) {
			select {
			case rd.keyChan <- k:
			case <-rd.quit:
			}
		} else if cpr != (pos{}) {
			select {
			case rd.cprChan <- cpr:
			case <-rd.quit:
			}
		} else if mouse != (mouseEvent{}) {
			select {
			case rd.mouseChan <- mouse:
			case <-rd.quit:
			}
		} else if paste != nil {
			select {
			case rd.pasteChan <- *paste:
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
			k = uitypes.Key{'[', uitypes.Ctrl}
			break
		}
		switch r2 {
		case '[':
			// A '[' follows. CSI style function key sequence.
			r = readRune()
			if r == runeTimeout || r == runeReadError {
				k = uitypes.Key{'[', uitypes.Alt}
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
				mouse = mouseEvent{
					pos{int(cy) - 32, int(cx) - 32}, down, button, mod}
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
				cpr = pos{nums[0], nums[1]}
			} else if starter == '<' && (r == 'm' || r == 'M') {
				// SGR-style mouse event.
				if len(nums) != 3 {
					badSeq("bad SGR mouse event")
					return
				}
				down := r == 'M'
				button := nums[0] & 3
				mod := mouseModify(nums[0])
				mouse = mouseEvent{pos{nums[2], nums[1]}, down, button, mod}
			} else if r == '~' && len(nums) == 1 && (nums[0] == 200 || nums[0] == 201) {
				b := nums[0] == 200
				paste = &b
			} else {
				k = parseCSI(nums, r, currentSeq)
				if k == (uitypes.Key{}) {
					badSeq("bad CSI")
				}
			}
		case 'O':
			// An 'O' follows. G3 style function key sequence: read one rune.
			r = readRune()
			if r == runeTimeout || r == runeReadError {
				// Nothing follows after 'O'. Taken as uitypes.Alt-o.
				k = uitypes.Key{'o', uitypes.Alt}
				return
			}
			r, ok := g3Seq[r]
			if ok {
				k = uitypes.Key{r, 0}
			} else {
				badSeq("bad G3")
			}
		default:
			// Something other than '[' or 'O' follows. Taken as an
			// uitypes.Alt-modified key, possibly also modified by uitypes.Ctrl.
			k = ctrlModify(r2)
			k.Mod |= uitypes.Alt
		}
	default:
		k = ctrlModify(r)
	}
}

// ctrlModify determines whether a rune corresponds to a uitypes.Ctrl-modified key and
// returns the uitypes.Key the rune represents.
func ctrlModify(r rune) uitypes.Key {
	switch r {
	case 0x0:
		return uitypes.Key{'`', uitypes.Ctrl} // ^@
	case 0x1e:
		return uitypes.Key{'6', uitypes.Ctrl} // ^^
	case 0x1f:
		return uitypes.Key{'/', uitypes.Ctrl} // ^_
	case uitypes.Tab, uitypes.Enter, uitypes.Backspace: // ^I ^J ^?
		return uitypes.Key{r, 0}
	default:
		// Regular uitypes.Ctrl sequences.
		if 0x1 <= r && r <= 0x1d {
			return uitypes.Key{r + 0x40, uitypes.Ctrl}
		}
	}
	return uitypes.Key{r, 0}
}

// G3-style key sequences: \eO followed by exactly one character. For instance,
// \eOP is uitypes.F1.
var g3Seq = map[rune]rune{
	'A': uitypes.Up, 'B': uitypes.Down, 'C': uitypes.Right, 'D': uitypes.Left,

	// uitypes.F1-uitypes.F4: xterm, libvte and tmux
	'P': uitypes.F1, 'Q': uitypes.F2,
	'R': uitypes.F3, 'S': uitypes.F4,

	// uitypes.Home and uitypes.End: libvte
	'H': uitypes.Home, 'F': uitypes.End,
}

// Tables for CSI-style key sequences, which are \e[ followed by a list of
// semicolon-delimited numeric arguments, before being concluded by a
// non-numeric, non-semicolon rune.

// CSI-style key sequences that can be identified based on the ending rune. For
// instance, \e[A is uitypes.Up.
var keyByLast = map[rune]uitypes.Key{
	'A': uitypes.Key{uitypes.Up, 0}, 'B': uitypes.Key{uitypes.Down, 0},
	'C': uitypes.Key{uitypes.Right, 0}, 'D': uitypes.Key{uitypes.Left, 0},
	'H': uitypes.Key{uitypes.Home, 0}, 'F': uitypes.Key{uitypes.End, 0},
	'Z': uitypes.Key{uitypes.Tab, uitypes.Shift},
}

// CSI-style key sequences ending with '~' and can be identified based on
// the only number argument. For instance, \e[1~ is uitypes.Home.
var keyByNum0 = map[int]rune{
	1: uitypes.Home, 2: uitypes.Insert, 3: uitypes.Delete, 4: uitypes.End,
	5: uitypes.PageUp, 6: uitypes.PageDown,
	11: uitypes.F1, 12: uitypes.F2, 13: uitypes.F3, 14: uitypes.F4,
	15: uitypes.F5, 17: uitypes.F6, 18: uitypes.F7, 19: uitypes.F8,
	20: uitypes.F9, 21: uitypes.F10, 23: uitypes.F11, 24: uitypes.F12,
}

// CSI-style key sequences ending with '~', with 27 as the first numeric
// argument. For instance, \e[27;9~ is uitypes.Tab.
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
func parseCSI(nums []int, last rune, seq string) uitypes.Key {
	if k, ok := keyByLast[last]; ok {
		if len(nums) == 0 {
			// Unmodified: \e[A (uitypes.Up)
			return k
		} else if len(nums) == 2 && nums[0] == 1 {
			// Modified: \e[1;5A (uitypes.Ctrl-uitypes.Up)
			return xtermModify(k, nums[1], seq)
		} else {
			return uitypes.Key{}
		}
	}

	if last == '~' {
		if len(nums) == 1 || len(nums) == 2 {
			if r, ok := keyByNum0[nums[0]]; ok {
				k := uitypes.Key{r, 0}
				if len(nums) == 1 {
					// Unmodified: \e[5~ (uitypes.PageUp)
					return k
				}
				// Modified: \e[5;5~ (uitypes.Ctrl-uitypes.PageUp)
				return xtermModify(k, nums[1], seq)
			}
		} else if len(nums) == 3 && nums[0] == 27 {
			if r, ok := keyByNum2[nums[2]]; ok {
				k := uitypes.Key{r, 0}
				return xtermModify(k, nums[1], seq)
			}
		}
	}

	return uitypes.Key{}
}

func xtermModify(k uitypes.Key, mod int, seq string) uitypes.Key {
	switch mod {
	case 0:
		// do nothing
	case 2:
		k.Mod |= uitypes.Shift
	case 3:
		k.Mod |= uitypes.Alt
	case 4:
		k.Mod |= uitypes.Shift | uitypes.Alt
	case 5:
		k.Mod |= uitypes.Ctrl
	case 6:
		k.Mod |= uitypes.Shift | uitypes.Ctrl
	case 7:
		k.Mod |= uitypes.Alt | uitypes.Ctrl
	case 8:
		k.Mod |= uitypes.Shift | uitypes.Alt | uitypes.Ctrl
	default:
		return uitypes.Key{}
	}
	return k
}

func mouseModify(n int) uitypes.Mod {
	var mod uitypes.Mod
	if n&4 != 0 {
		mod |= uitypes.Shift
	}
	if n&8 != 0 {
		mod |= uitypes.Alt
	}
	if n&16 != 0 {
		mod |= uitypes.Ctrl
	}
	return mod
}

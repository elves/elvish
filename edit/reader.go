package edit

import (
	"fmt"
	"os"
	"time"
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

// Reader converts a stream of runes into a stream of OneRead's.
type Reader struct {
	ar        *AsyncReader
	keyChan   chan Key
	cprChan   chan pos
	mouseChan chan mouseEvent
	errChan   chan error
	quit      chan struct{}
}

type mouseEvent struct {
	pos
	down   bool
	button int
	mod    Mod
}

// BadEscSeq indicates that a escape sequence has been read from the terminal,
// but it cannot be parsed.
type BadEscSeq struct {
	seq string
	msg string
}

func newBadEscSeq(seq string, msg string) *BadEscSeq {
	return &BadEscSeq{seq, msg}
}

func (bes *BadEscSeq) Error() string {
	return fmt.Sprintf("bad escape sequence %q: %s", bes.seq, bes.msg)
}

// NewReader creates a new Reader on the given terminal file.
func NewReader(f *os.File) *Reader {
	rd := &Reader{
		NewAsyncReader(f),
		make(chan Key),
		make(chan pos),
		make(chan mouseEvent),
		make(chan error),
		nil,
	}
	return rd
}

// KeyChan returns the channel onto which the Reader writes Keys it has read.
func (rd *Reader) KeyChan() <-chan Key {
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

// ErrorChan returns the channel onto which the Reader writes errors it came
// across during the reading process.
func (rd *Reader) ErrorChan() <-chan error {
	return rd.errChan
}

// Run runs the Reader. It blocks until Quit is called and should be called in
// a separate goroutine.
func (rd *Reader) Run() {
	runes := rd.ar.Chan()
	rd.quit = make(chan struct{})
	go rd.ar.Run()

	for {
		select {
		case r := <-runes:
			rd.readOne(r)
		case <-rd.quit:
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
	var k Key
	var cpr pos
	var mouse mouseEvent
	var err error
	var currentSeq string

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
		if k != (Key{}) {
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
		}
		if err != nil {
			select {
			case rd.errChan <- err:
			case <-rd.quit:
			}
		}
	}()

	switch r {
	case Tab, Enter, Backspace:
		k = Key{r, 0}
	case 0x0:
		k = Key{'`', Ctrl} // ^@
	case 0x1d:
		k = Key{'6', Ctrl} // ^^
	case 0x1f:
		k = Key{'/', Ctrl} // ^_
	case 0x1b: // ^[ Escape
		r2 := readRune()
		if r2 == runeTimeout || r2 == runeReadError {
			k = Key{'[', Ctrl}
			break
		}
		switch r2 {
		case '[':
			// CSI style function key sequence, looks like [\d;]*[^\d;]
			// Read numeric parameters (if any)
			nums := make([]int, 0, 2)
			isMouse := false
		CSISeq:
			for {
				r = readRune()
				// Timeout can only happen at first readRune.
				if r == runeTimeout || r == runeReadError {
					k = Key{'[', Alt}
					return
				}
				switch {
				case r == ';':
					nums = append(nums, 0)
				case r == '<':
					isMouse = true
				case '0' <= r && r <= '9':
					if len(nums) == 0 {
						nums = append(nums, 0)
					}
					cur := len(nums) - 1
					nums[cur] = nums[cur]*10 + int(r-'0')
				default:
					break CSISeq
				}
			}
			if r == 'R' {
				// CPR
				if len(nums) != 2 {
					err = newBadEscSeq(currentSeq, "bad CPR")
					return
				}
				cpr = pos{nums[0], nums[1]}
			} else if isMouse && (r == 'm' || r == 'M') {
				if len(nums) != 3 {
					err = newBadEscSeq(currentSeq, "bad mouse event")
					return
				}
				down := r == 'M'
				n0 := nums[0]
				button := n0 & 3
				mod := Mod(0)
				if n0&4 != 0 {
					mod |= Shift
				}
				if n0&8 != 0 {
					mod |= Alt
				}
				if n0&16 != 0 {
					mod |= Ctrl
				}
				mouse = mouseEvent{pos{nums[2], nums[1]}, down, button, mod}
			} else {
				k = parseCSI(nums, r, currentSeq)
				if k == (Key{}) {
					err = newBadEscSeq(currentSeq, "bad CSI")
				}
			}
		case 'O':
			// G3 style function key sequence: read one rune.
			r = readRune()
			if r == runeTimeout || r == runeReadError {
				k = Key{r2, Alt}
				return
			}
			r, ok := g3Seq[r]
			if ok {
				k = Key{r, 0}
			} else {
				err = newBadEscSeq(currentSeq, "bad G3")
			}
		default:
			k = Key{r2, Alt}
		}
	default:
		// Regular Ctrl sequences.
		if 0x1 <= r && r <= 0x1d {
			k = Key{r + 0x40, Ctrl}
		} else {
			k = Key{r, 0}
		}
	}
}

// G3-style key sequences: \eO followed by exactly one character. For instance,
// \eOP is F1.
var g3Seq = map[rune]rune{
	// F1-F4: xterm, libvte and tmux
	'P': F1, 'Q': F2,
	'R': F3, 'S': F4,

	// Home and End: libvte
	'H': Home, 'F': End,
}

// Tables for CSI-style key sequences, which are \e[ followed by a list of
// semicolon-delimited numeric arguments, before being concluded by a
// non-numeric, non-semicolon rune.

// CSI-style key sequences that can be identified based on the ending rune. For
// instance, \e[A is Up.
var keyByLast = map[rune]Key{
	'A': Key{Up, 0}, 'B': Key{Down, 0},
	'C': Key{Right, 0}, 'D': Key{Left, 0},
	'H': Key{Home, 0}, 'F': Key{End, 0},
	'Z': Key{Tab, Shift},
}

// CSI-style key sequences ending with '~' and can be identified based on
// the only number argument. For instance, \e[1~ is Home.
var keyByNum0 = map[int]rune{
	1: Home, 2: Insert, 3: Delete, 4: End, 5: PageUp, 6: PageDown,
	11: F1, 12: F2, 13: F3, 14: F4,
	15: F5, 17: F6, 18: F7, 19: F8, 20: F9, 21: F10, 23: F11, 24: F12,
}

// CSI-style key sequences ending with '~', with 27 as the first numeric
// argument. For instance, \e[27;9~ is Tab.
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
func parseCSI(nums []int, last rune, seq string) Key {
	if k, ok := keyByLast[last]; ok {
		if len(nums) == 0 {
			// Unmodified: \e[A (Up)
			return k
		} else if len(nums) == 2 && nums[0] == 1 {
			// Modified: \e[1;5A (Ctrl-Up)
			return xtermModify(k, nums[1], seq)
		} else {
			return Key{}
		}
	}

	if last == '~' {
		if len(nums) == 1 || len(nums) == 2 {
			if r, ok := keyByNum0[nums[0]]; ok {
				k := Key{r, 0}
				if len(nums) == 1 {
					// Unmodified: \e[5~ (PageUp)
					return k
				}
				// Modified: \e[5;5~ (Ctrl-PageUp)
				return xtermModify(k, nums[1], seq)
			}
		} else if len(nums) == 3 && nums[0] == 27 {
			if r, ok := keyByNum2[nums[2]]; ok {
				k := Key{r, 0}
				return xtermModify(k, nums[1], seq)
			}
		}
	}

	return Key{}
}

func xtermModify(k Key, mod int, seq string) Key {
	switch mod {
	case 0:
		// do nothing
	case 2:
		k.Mod |= Shift
	case 3:
		k.Mod |= Alt
	case 4:
		k.Mod |= Shift | Alt
	case 5:
		k.Mod |= Ctrl
	case 6:
		k.Mod |= Shift | Ctrl
	case 7:
		k.Mod |= Alt | Ctrl
	case 8:
		k.Mod |= Shift | Alt | Ctrl
	default:
		return Key{}
	}
	return k
}

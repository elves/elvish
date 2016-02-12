package edit

import (
	"fmt"
	"os"
	"time"

	"github.com/elves/elvish/errutil"
)

var (
	// EscTimeout is the amount of time after which an \033 (\x1b) character is
	// parsed as a standalone Esc key press instead of starting an escape
	// sequence. Modern terminal emulators send escape sequences very fast, so
	// 10ms is more than sufficient. SSH connections on a slow link can be
	// problematic,
	EscTimeout = 10 * time.Millisecond
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
	ar         *AsyncReader
	ones       chan OneRead
	quit       chan struct{}
	currentSeq string
	readError  error
}

// OneRead is a single unit that is read, which is either a key, a CPR, or an
// error.
type OneRead struct {
	Key Key
	CPR pos
	Err error
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

func (rd *Reader) badEscSeq(msg string) {
	errutil.Throw(newBadEscSeq(rd.currentSeq, msg))
}

// NewReader creates a new Reader on the given terminal file.
func NewReader(f *os.File) *Reader {
	rd := &Reader{
		ar:   NewAsyncReader(f),
		ones: make(chan OneRead),
		quit: make(chan struct{}),
	}
	return rd
}

// Chan returns the channel onto which the Reader writes OneRead's.
func (rd *Reader) Chan() <-chan OneRead {
	return rd.ones
}

// Run runs the Reader. It blocks until Quit is called and should be called in
// a separate goroutine.
func (rd *Reader) Run() {
	runes := rd.ar.Chan()
	go rd.ar.Run()

	for {
		select {
		case r := <-runes:
			k, c, e := rd.readOne(r)
			select {
			case rd.ones <- OneRead{k, c, e}:
			case <-rd.quit:
				return
			}
		case <-rd.quit:
			return
		}
	}
}

// Quit terminates the loop of Run.
func (rd *Reader) Quit() {
	rd.ar.Quit()
	rd.quit <- struct{}{}
}

// Close releases files and channels associated with the Reader. It does not
// close the file used to create it.
func (rd *Reader) Close() {
	rd.ar.Close()
	close(rd.ones)
	close(rd.quit)
}

// readOne attempts to read one key or CPR, led by a rune already read.
func (rd *Reader) readOne(r rune) (k Key, cpr pos, err error) {
	defer errutil.Catch(&err)

	rd.currentSeq = ""

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
		//rd.timed.Timeout = escTimeout
		//defer func() { rd.timed.Timeout = -1 }()
		r2 := rd.readRune(EscTimeout)
		if r2 == runeTimeout {
			return Key{'[', Ctrl}, invalidPos, nil
		} else if r2 == runeReadError {
			return Key{'[', Ctrl}, invalidPos, rd.readError
		}
		switch r2 {
		case '[':
			// CSI style function key sequence, looks like [\d;]*[^\d;]
			// Read numeric parameters (if any)
			nums := make([]int, 0, 2)
			seq := "\x1b["
			timeout := EscTimeout
			for {
				r = rd.readRune(timeout)
				// Timeout can only happen at first readRune.
				if r == runeTimeout {
					return Key{'[', Alt}, invalidPos, nil
				} else if r == runeReadError {
					return Key{'[', Alt}, invalidPos, rd.readError
				}
				seq += string(r)
				// After first rune read we turn off the timeout
				timeout = -1
				if r != ';' && (r < '0' || r > '9') {
					break
				}

				if len(nums) == 0 {
					nums = append(nums, 0)
				}
				if r == ';' {
					nums = append(nums, 0)
				} else {
					cur := len(nums) - 1
					nums[cur] = nums[cur]*10 + int(r-'0')
				}
			}
			if r == 'R' {
				// CPR
				if len(nums) != 2 {
					rd.badEscSeq("bad cpr")
				}
				return ZeroKey, pos{nums[0], nums[1]}, nil
			}
			k, err := parseCSI(nums, r, seq)
			return k, invalidPos, err
		case 'O':
			// G3 style function key sequence: read one rune.
			r = rd.readRune(EscTimeout)
			if r == runeTimeout {
				return Key{r2, Alt}, invalidPos, nil
			} else if r == runeReadError {
				return Key{r2, Alt}, invalidPos, rd.readError
			}
			r, ok := g3Seq[r]
			if ok {
				return Key{r, 0}, invalidPos, nil
			}
			rd.badEscSeq("")
		}
		return Key{r2, Alt}, invalidPos, nil
	default:
		// Sane Ctrl- sequences that agree with the keyboard...
		if 0x1 <= r && r <= 0x1d {
			k = Key{r + 0x40, Ctrl}
		} else {
			k = Key{r, 0}
		}
	}
	return k, invalidPos, nil
}

func (rd *Reader) readRune(d time.Duration) rune {
	select {
	case r := <-rd.ar.Chan():
		rd.currentSeq += string(r)
		return r
	case err := <-rd.ar.ErrorChan():
		rd.readError = err
		return runeReadError
	case <-After(d):
		return runeTimeout
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
func parseCSI(nums []int, last rune, seq string) (Key, error) {
	if k, ok := keyByLast[last]; ok {
		if len(nums) == 0 {
			// Unmodified: \e[A (Up)
			return k, nil
		} else if len(nums) == 2 && nums[0] == 1 {
			// Modified: \e[1;5A (Ctrl-Up)
			return xtermModify(k, nums[1], seq)
		} else {
			return ZeroKey, newBadEscSeq(seq, "")
		}
	}

	if last == '~' {
		if len(nums) == 1 || len(nums) == 2 {
			if r, ok := keyByNum0[nums[0]]; ok {
				k := Key{r, 0}
				if len(nums) == 1 {
					// Unmodified: \e[5~ (PageUp)
					return k, nil
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

	return ZeroKey, newBadEscSeq(seq, "")
}

func xtermModify(k Key, mod int, seq string) (Key, error) {
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
		return ZeroKey, newBadEscSeq(seq, "")
	}
	return k, nil
}

package edit

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/elves/elvish/util"
)

const (
	ReaderOutChanSize int = 16
)

// readerCtrl is used to control the internal reader goroutine.
type readerCtrl byte

// Possible values for readerCtrl.
const (
	readerStop readerCtrl = iota
	readerContinue
	readerQuit
)

const (
	EscTimeout time.Duration = 10 * time.Millisecond
	CPRTimeout               = 10 * time.Millisecond
)

const (
	RuneTimeout rune = -1
)

var (
	ErrTimeout = errors.New("timed out")
	BadCPR     = errors.New("bad CPR")
)

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

type OneRead struct {
	Key Key
	CPR pos
	Err error
}

// Reader converts a stream of runes into a stream of Keys
type Reader struct {
	ar         *util.AsyncReader
	ones       chan OneRead
	ctrl       chan readerCtrl
	ctrlAck    chan bool
	currentSeq string
}

func NewReader(f *os.File) *Reader {
	rd := &Reader{
		ar:      util.NewAsyncReader(f),
		ones:    make(chan OneRead, ReaderOutChanSize),
		ctrl:    make(chan readerCtrl),
		ctrlAck: make(chan bool),
	}
	go rd.run()
	return rd
}

func (rd *Reader) Chan() <-chan OneRead {
	return rd.ones
}

func (rd *Reader) sendCtrl(c readerCtrl) {
	rd.ctrl <- c
	<-rd.ctrlAck
}

func (rd *Reader) Stop() {
	rd.ar.Stop()
	rd.sendCtrl(readerStop)
}

func (rd *Reader) Continue() {
	rd.ar.Continue()
	rd.sendCtrl(readerContinue)
}

func (rd *Reader) Quit() {
	rd.ar.Quit()
	rd.sendCtrl(readerQuit)
}

func (rd *Reader) badEscSeq(msg string) {
	util.Throw(newBadEscSeq(rd.currentSeq, msg))
}

func (rd *Reader) readRune(d time.Duration) rune {
	select {
	case r := <-rd.ar.Chan():
		rd.currentSeq += string(r)
		return r
	case <-util.After(d):
		return RuneTimeout
	}
}

func (rd *Reader) readAssertedRune(r rune, d time.Duration) {
	if rd.readRune(d) != r {
		rd.badEscSeq("Expect " + string(r))
	}
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

// G3 style function key sequences: ^[O followed by exactly one character.
var g3Seq = map[rune]rune{
	// F1-F4: xterm, libvte and tmux
	'P': F1, 'Q': F2,
	'R': F3, 'S': F4,

	// Home and End: libvte
	'H': Home, 'F': End,
}

func (rd *Reader) readOne(r rune) (k Key, cpr pos, err error) {
	defer util.Catch(&err)

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
		//rd.timed.Timeout = EscTimeout
		//defer func() { rd.timed.Timeout = -1 }()
		r2 := rd.readRune(EscTimeout)
		if r2 == RuneTimeout {
			return Key{'[', Ctrl}, InvalidPos, nil
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
				if r == RuneTimeout {
					return Key{'[', Alt}, InvalidPos, nil
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
			} else {
				k, err := parseCSI(nums, r, seq)
				return k, InvalidPos, err
			}
		case 'O':
			// G3 style function key sequence: read one rune.
			r = rd.readRune(EscTimeout)
			if r == RuneTimeout {
				return Key{r2, Alt}, InvalidPos, nil
			}
			r, ok := g3Seq[r]
			if ok {
				return Key{r, 0}, InvalidPos, nil
			}
			rd.badEscSeq("")
		}
		return Key{r2, Alt}, InvalidPos, nil
	default:
		// Sane Ctrl- sequences that agree with the keyboard...
		if 0x1 <= r && r <= 0x1d {
			k = Key{r + 0x40, Ctrl}
		} else {
			k = Key{r, 0}
		}
	}
	return k, InvalidPos, nil
}

func (rd *Reader) stop() (quit bool) {
	for {
		select {
		case ctrl := <-rd.ctrl:
			rd.ctrlAck <- true
			switch ctrl {
			case readerQuit:
				return true
			case readerContinue:
				return false
			}
		}
	}
}

func (rd *Reader) run() {
	defer close(rd.ones)

	runes := rd.ar.Chan()

	for {
		select {
		case r := <-runes:
			k, c, e := rd.readOne(r)
			rd.ones <- OneRead{k, c, e}
		case ctrl := <-rd.ctrl:
			rd.ctrlAck <- true
			switch ctrl {
			case readerStop:
				if rd.stop() {
					return
				}
			case readerQuit:
				return
			}
		}
	}
}

var keyByLast = map[rune]Key{
	'A': Key{Up, 0}, 'B': Key{Down, 0},
	'C': Key{Right, 0}, 'D': Key{Left, 0},
	'H': Key{Home, 0}, 'F': Key{End, 0},
	'Z': Key{Tab, Shift},
}

// last == '~'
var keyByNum0 = map[int]rune{
	1: Home, 2: Insert, 3: Delete, 4: End, 5: PageUp, 6: PageDown,
	11: F1, 12: F2, 13: F3, 14: F4,
	15: F5, 17: F6, 18: F7, 19: F8, 20: F9, 21: F10, 23: F11, 24: F12,
}

// last == '~', num[0] == 27
// The list is taken blindly from tmux source xterm-keys.c. I don't have a
// keyboard that can generate such sequences, but assumably some PC keyboard
// with a numpad can.
var keyByNum2 = map[int]rune{
	9: '\t', 13: '\r',
	33: '!', 35: '#', 39: '\'', 40: '(', 41: ')', 43: '+', 44: ',', 45: '-',
	46: '.',
	48: '0', 49: '1', 50: '2', 51: '3', 52: '4', 53: '5', 54: '6', 55: '7',
	56: '8', 57: '9',
	58: ':', 59: ';', 60: '<', 61: '=', 62: '>', 63: ';',
}

// Parse a CSI-style function key sequence.
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

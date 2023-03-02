//go:build unix

package term

import (
	"os"
	"time"

	"src.elv.sh/pkg/ui"
)

// reader reads terminal escape sequences and decodes them into events.
type reader struct {
	fr fileReader
}

func newReader(f *os.File) *reader {
	fr, err := newFileReader(f)
	if err != nil {
		// TODO(xiaq): Do not panic.
		panic(err)
	}
	return &reader{fr}
}

func (rd *reader) ReadEvent() (Event, error) {
	return readEvent(rd.fr)
}

func (rd *reader) ReadRawEvent() (Event, error) {
	r, err := readRune(rd.fr, -1)
	return K(r), err
}

func (rd *reader) Close() {
	rd.fr.Stop()
	rd.fr.Close()
}

// Used by readRune in readOne to signal end of current sequence.
const runeEndOfSeq rune = -1

// Timeout for bytes in escape sequences. Modern terminal emulators send escape
// sequences very fast, so 10ms is more than sufficient. SSH connections on a
// slow link might be problematic though.
var keySeqTimeout = 10 * time.Millisecond

func readEvent(rd byteReaderWithTimeout) (event Event, err error) {
	var r rune
	r, err = readRune(rd, -1)
	if err != nil {
		return
	}

	currentSeq := string(r)
	// Attempts to read a rune within a timeout of keySeqTimeout. It returns
	// runeEndOfSeq if there is any error; the caller should terminate the
	// current sequence when it sees that value.
	readRune :=
		func() rune {
			r, e := readRune(rd, keySeqTimeout)
			if e != nil {
				return runeEndOfSeq
			}
			currentSeq += string(r)
			return r
		}
	badSeq := func(msg string) {
		err = seqError{msg, currentSeq}
	}

	switch r {
	case 0x1b: // ^[ Escape
		r2 := readRune()
		// According to https://unix.stackexchange.com/a/73697, rxvt and derivatives
		// prepend another ESC to a CSI-style or G3-style sequence to signal Alt.
		// If that happens, remember this now; it will be later picked up when parsing
		// those two kinds of sequences.
		//
		// issue #181
		hasTwoLeadingESC := false
		if r2 == 0x1b {
			hasTwoLeadingESC = true
			r2 = readRune()
		}
		if r2 == runeEndOfSeq {
			// TODO(xiaq): Error is swallowed.
			// Nothing follows. Taken as a lone Escape.
			event = KeyEvent{'[', ui.Ctrl}
			break
		}
		switch r2 {
		case '[':
			// A '[' follows. CSI style function key sequence.
			r = readRune()
			if r == runeEndOfSeq {
				event = KeyEvent{'[', ui.Alt}
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
				if cb == runeEndOfSeq {
					badSeq("incomplete mouse event")
					return
				}
				cx := readRune()
				if cx == runeEndOfSeq {
					badSeq("incomplete mouse event")
					return
				}
				cy := readRune()
				if cy == runeEndOfSeq {
					badSeq("incomplete mouse event")
					return
				}
				down := true
				button := int(cb & 3)
				if button == 3 {
					down = false
					button = -1
				}
				mod := mouseModify(int(cb))
				event = MouseEvent{
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
				case r == runeEndOfSeq:
					// Incomplete CSI.
					badSeq("incomplete CSI")
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
				event = CursorPosition{nums[0], nums[1]}
			} else if starter == '<' && (r == 'm' || r == 'M') {
				// SGR-style mouse event.
				if len(nums) != 3 {
					badSeq("bad SGR mouse event")
					return
				}
				down := r == 'M'
				button := nums[0] & 3
				mod := mouseModify(nums[0])
				event = MouseEvent{Pos{nums[2], nums[1]}, down, button, mod}
			} else if r == '~' && len(nums) == 1 && (nums[0] == 200 || nums[0] == 201) {
				b := nums[0] == 200
				event = PasteSetting(b)
			} else {
				k := parseCSI(nums, r, currentSeq)
				if k == (ui.Key{}) {
					badSeq("bad CSI")
				} else {
					if hasTwoLeadingESC {
						k.Mod |= ui.Alt
					}
					event = KeyEvent(k)
				}
			}
		case 'O':
			// An 'O' follows. G3 style function key sequence: read one rune.
			r = readRune()
			if r == runeEndOfSeq {
				// Nothing follows after 'O'. Taken as Alt-O.
				event = KeyEvent{'O', ui.Alt}
				return
			}
			k, ok := g3Seq[r]
			if ok {
				if hasTwoLeadingESC {
					k.Mod |= ui.Alt
				}
				event = KeyEvent(k)
			} else {
				badSeq("bad G3")
			}
		default:
			// Something other than '[' or 'O' follows. Taken as an
			// Alt-modified key, possibly also modified by Ctrl.
			k := ctrlModify(r2)
			k.Mod |= ui.Alt
			event = KeyEvent(k)
		}
	default:
		event = KeyEvent(ctrlModify(r))
	}
	return
}

// Determines whether a rune corresponds to a Ctrl-modified key and returns the
// ui.Key the rune represents.
func ctrlModify(r rune) ui.Key {
	switch r {
	// TODO(xiaq): Are the following special cases universal?
	case 0x0:
		return ui.K('`', ui.Ctrl) // ^@
	case 0x1e:
		return ui.K('6', ui.Ctrl) // ^^
	case 0x1f:
		return ui.K('/', ui.Ctrl) // ^_
	case ui.Tab, ui.Enter, ui.Backspace: // ^I ^J ^?
		// Ambiguous Ctrl keys; prefer the non-Ctrl form as they are more likely.
		return ui.K(r)
	default:
		// Regular ui.Ctrl sequences.
		if 0x1 <= r && r <= 0x1d {
			return ui.K(r+0x40, ui.Ctrl)
		}
	}
	return ui.K(r)
}

// Tables for key sequences. Comments document which terminal emulators are
// known to generate which sequences. The terminal emulators tested are
// categorized into xterm (including actual xterm, libvte-based terminals,
// Konsole and Terminal.app unless otherwise noted), urxvt, tmux.

// G3-style key sequences: \eO followed by exactly one character. For instance,
// \eOP is F1. These are pretty limited in that they cannot be extended to
// support modifier keys, other than a leading \e for Alt (e.g. \e\eOP is
// Alt-F1). Terminals that send G3-style key sequences typically switch to
// sending a CSI-style key sequence when a non-Alt modifier key is pressed.
var g3Seq = map[rune]ui.Key{
	// xterm, tmux -- only in Vim, depends on termios setting?
	// NOTE(xiaq): According to urxvt's manpage, \eO[ABCD] sequences are used for
	// Ctrl-Shift-modified arrow keys; however, this doesn't seem to be true for
	// urxvt 9.22 packaged by Debian; those keys simply send the same sequence
	// as Ctrl-modified keys (\eO[abcd]).
	'A': ui.K(ui.Up), 'B': ui.K(ui.Down), 'C': ui.K(ui.Right), 'D': ui.K(ui.Left),
	'H': ui.K(ui.Home), 'F': ui.K(ui.End), 'M': ui.K(ui.Insert),
	// urxvt
	'a': ui.K(ui.Up, ui.Ctrl), 'b': ui.K(ui.Down, ui.Ctrl),
	'c': ui.K(ui.Right, ui.Ctrl), 'd': ui.K(ui.Left, ui.Ctrl),
	// xterm, urxvt, tmux
	'P': ui.K(ui.F1), 'Q': ui.K(ui.F2), 'R': ui.K(ui.F3), 'S': ui.K(ui.F4),
}

// Tables for CSI-style key sequences. A CSI sequence is \e[ followed by zero or
// more numerical arguments (separated by semicolons), ending in a non-numeric,
// non-semicolon rune. They are used for many purposes, and CSI-style key
// sequences are a subset of them.
//
// There are several variants of CSI-style key sequences; see comments above the
// respective tables. In all variants, modifier keys are encoded in numerical
// arguments; see xtermModify. Note that although the set of possible sequences
// make it possible to express a very complete set of key combinations, they are
// not always sent by terminals. For instance, many (if not most) terminals will
// send the same sequence for Up when Shift-Up is pressed, even if Shift-Up is
// expressible using the escape sequences described below.

// CSI-style key sequences identified by the last rune. For instance, \e[A is
// Up. When modified, two numerical arguments are added, the first always being
// 1 and the second identifying the modifier. For instance, \e[1;5A is Ctrl-Up.
var csiSeqByLast = map[rune]ui.Key{
	// xterm, urxvt, tmux
	'A': ui.K(ui.Up), 'B': ui.K(ui.Down), 'C': ui.K(ui.Right), 'D': ui.K(ui.Left),
	// urxvt
	'a': ui.K(ui.Up, ui.Shift), 'b': ui.K(ui.Down, ui.Shift),
	'c': ui.K(ui.Right, ui.Shift), 'd': ui.K(ui.Left, ui.Shift),
	// xterm (Terminal.app only sends those in alternate screen)
	'H': ui.K(ui.Home), 'F': ui.K(ui.End),
	// xterm, urxvt, tmux
	'Z': ui.K(ui.Tab, ui.Shift),
}

// CSI-style key sequences ending with '~' with by one or two numerical
// arguments. The first argument identifies the key, and the optional second
// argument identifies the modifier. For instance, \e[3~ is Delete, and \e[3;5~
// is Ctrl-Delete.
//
// An alternative encoding of the modifier key, only known to be used by urxvt
// (or for that matter, likely also rxvt) is to change the last rune: '$' for
// Shift, '^' for Ctrl, and '@' for Ctrl+Shift. The numeric argument is kept
// unchanged. For instance, \e[3^ is Ctrl-Delete.
var csiSeqTilde = map[int]rune{
	// tmux (NOTE: urxvt uses the pair for Find/Select)
	1: ui.Home, 4: ui.End,
	// xterm (Terminal.app sends ^M for Fn+Enter), urxvt, tmux
	2: ui.Insert,
	// xterm, urxvt, tmux
	3: ui.Delete,
	// xterm (Terminal.app only sends those in alternate screen), urxvt, tmux
	// NOTE: called Prior/Next in urxvt manpage
	5: ui.PageUp, 6: ui.PageDown,
	// urxvt
	7: ui.Home, 8: ui.End,
	// urxvt
	11: ui.F1, 12: ui.F2, 13: ui.F3, 14: ui.F4,
	// xterm, urxvt, tmux
	// NOTE: 16 and 22 are unused
	15: ui.F5, 17: ui.F6, 18: ui.F7, 19: ui.F8,
	20: ui.F9, 21: ui.F10, 23: ui.F11, 24: ui.F12,
}

// CSI-style key sequences ending with '~', with the first argument always 27,
// the second argument identifying the modifier, and the third argument
// identifying the key. For instance, \e[27;5;9~ is Ctrl-Tab.
//
// NOTE(xiaq): The list is taken blindly from xterm-keys.c in the tmux source
// tree. I do not have a keyboard-terminal combination that generate such
// sequences, but assumably they are generated by some terminals for numpad
// inputs.
var csiSeqTilde27 = map[int]rune{
	9: '\t', 13: '\r',
	33: '!', 35: '#', 39: '\'', 40: '(', 41: ')', 43: '+', 44: ',', 45: '-',
	46: '.',
	48: '0', 49: '1', 50: '2', 51: '3', 52: '4', 53: '5', 54: '6', 55: '7',
	56: '8', 57: '9',
	58: ':', 59: ';', 60: '<', 61: '=', 62: '>', 63: ';',
}

// parseCSI parses a CSI-style key sequence. See comments above for all the 3
// variants this function handles.
func parseCSI(nums []int, last rune, seq string) ui.Key {
	if k, ok := csiSeqByLast[last]; ok {
		if len(nums) == 0 {
			// Unmodified: \e[A (Up)
			return k
		} else if len(nums) == 2 && nums[0] == 1 {
			// Modified: \e[1;5A (Ctrl-Up)
			return xtermModify(k, nums[1], seq)
		} else {
			return ui.Key{}
		}
	}

	switch last {
	case '~':
		if len(nums) == 1 || len(nums) == 2 {
			if r, ok := csiSeqTilde[nums[0]]; ok {
				k := ui.K(r)
				if len(nums) == 1 {
					// Unmodified: \e[5~ (e.g. PageUp)
					return k
				}
				// Modified: \e[5;5~ (e.g. Ctrl-PageUp)
				return xtermModify(k, nums[1], seq)
			}
		} else if len(nums) == 3 && nums[0] == 27 {
			if r, ok := csiSeqTilde27[nums[2]]; ok {
				k := ui.K(r)
				return xtermModify(k, nums[1], seq)
			}
		}
	case '$', '^', '@':
		// Modified by urxvt; see comment above csiSeqTilde.
		if len(nums) == 1 {
			if r, ok := csiSeqTilde[nums[0]]; ok {
				var mod ui.Mod
				switch last {
				case '$':
					mod = ui.Shift
				case '^':
					mod = ui.Ctrl
				case '@':
					mod = ui.Shift | ui.Ctrl
				}
				return ui.K(r, mod)
			}
		}
	}

	return ui.Key{}
}

func xtermModify(k ui.Key, mod int, seq string) ui.Key {
	if mod < 0 || mod > 16 {
		// Out of range
		return ui.Key{}
	}
	if mod == 0 {
		return k
	}
	modFlags := mod - 1
	if modFlags&0x1 != 0 {
		k.Mod |= ui.Shift
	}
	if modFlags&0x2 != 0 {
		k.Mod |= ui.Alt
	}
	if modFlags&0x4 != 0 {
		k.Mod |= ui.Ctrl
	}
	if modFlags&0x8 != 0 {
		// This should be Meta, but we currently conflate Meta and Alt.
		k.Mod |= ui.Alt
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

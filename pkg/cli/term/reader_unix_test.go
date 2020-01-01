// +build !windows,!plan9

package term

import (
	"testing"

	"github.com/elves/elvish/pkg/ui"
)

var readEventTests = []struct {
	input string
	want  Event
}{
	// Simple graphical key.
	{"x", K('x')},
	{"X", K('X')},
	{" ", K(' ')},

	// Ctrl key.
	{"\001", K('A', ui.Ctrl)},
	{"\033", K('[', ui.Ctrl)},

	// Special Ctrl keys that do not obey the usual 0x40 rule.
	{"\000", K('`', ui.Ctrl)},
	{"\x1e", K('6', ui.Ctrl)},
	{"\x1f", K('/', ui.Ctrl)},

	// Ambiguous Ctrl keys; the reader uses the non-Ctrl form as canonical.
	{"\n", K('\n')},
	{"\t", K('\t')},
	{"\x7f", K('\x7f')}, // backspace

	// Alt plus simple graphical key.
	{"\033a", K('a', ui.Alt)},
	{"\033[", K('[', ui.Alt)},

	// G3-style key.
	{"\033OA", K(ui.Up)},
	{"\033OH", K(ui.Home)},

	// G3-style key with leading Escape.
	{"\033\033OA", K(ui.Up, ui.Alt)},
	{"\033\033OH", K(ui.Home, ui.Alt)},

	// Alt-O. This is handled as a special case because it looks like a G3-style
	// key.
	{"\033O", K('O', ui.Alt)},

	// CSI-sequence key identified by the ending rune.
	{"\033[A", K(ui.Up)},
	{"\033[H", K(ui.Home)},
	// Modifiers.
	{"\033[1;1A", K(ui.Up)},
	{"\033[1;2A", K(ui.Up, ui.Shift)},
	{"\033[1;3A", K(ui.Up, ui.Alt)},
	{"\033[1;4A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;5A", K(ui.Up, ui.Ctrl)},
	{"\033[1;6A", K(ui.Up, ui.Shift, ui.Ctrl)},
	{"\033[1;7A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;8A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},
	// The modifiers below should be for Meta, but we conflate Alt and Meta.
	{"\033[1;9A", K(ui.Up, ui.Alt)},
	{"\033[1;10A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;11A", K(ui.Up, ui.Alt)},
	{"\033[1;12A", K(ui.Up, ui.Shift, ui.Alt)},
	{"\033[1;13A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;14A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},
	{"\033[1;15A", K(ui.Up, ui.Alt, ui.Ctrl)},
	{"\033[1;16A", K(ui.Up, ui.Shift, ui.Alt, ui.Ctrl)},

	// CSI-sequence key with one argument, ending in '~'.
	{"\033[1~", K(ui.Home)},
	{"\033[11~", K(ui.F1)},
	// Modified.
	{"\033[1;2~", K(ui.Home, ui.Shift)},
	// Urxvt-flavor modifier, shifting the '~' to reflect the modifier
	{"\033[1$", K(ui.Home, ui.Shift)},
	{"\033[1^", K(ui.Home, ui.Ctrl)},
	{"\033[1@", K(ui.Home, ui.Shift, ui.Ctrl)},
	// With a leading Escape.
	{"\033\033[1~", K(ui.Home, ui.Alt)},

	// CSI-sequence key with three arguments and ending in '~'. The first
	// argument is always 27, the second identifies the modifier and the last
	// identifies the key.
	{"\033[27;4;63~", K(';', ui.Shift, ui.Alt)},

	// Cursor Position Report.
	{"\033[3;4R", CursorPosition{3, 4}},

	// Paste setting.
	{"\033[200~", PasteSetting(true)},
	{"\033[201~", PasteSetting(false)},

	// Mouse event.
	{"\033[M\x00\x23\x24", MouseEvent{Pos{4, 3}, true, 0, 0}},
	// Other buttons.
	{"\033[M\x01\x23\x24", MouseEvent{Pos{4, 3}, true, 1, 0}},
	// Button up.
	{"\033[M\x03\x23\x24", MouseEvent{Pos{4, 3}, false, -1, 0}},
	// Modified.
	{"\033[M\x04\x23\x24", MouseEvent{Pos{4, 3}, true, 0, ui.Shift}},
	{"\033[M\x08\x23\x24", MouseEvent{Pos{4, 3}, true, 0, ui.Alt}},
	{"\033[M\x10\x23\x24", MouseEvent{Pos{4, 3}, true, 0, ui.Ctrl}},
	{"\033[M\x14\x23\x24", MouseEvent{Pos{4, 3}, true, 0, ui.Shift | ui.Ctrl}},

	// SGR-style mouse event.
	{"\033[<0;3;4M", MouseEvent{Pos{4, 3}, true, 0, 0}},
	// Other buttons.
	{"\033[<1;3;4M", MouseEvent{Pos{4, 3}, true, 1, 0}},
	// Button up.
	{"\033[<0;3;4m", MouseEvent{Pos{4, 3}, false, 0, 0}},
	// Modified.
	{"\033[<4;3;4M", MouseEvent{Pos{4, 3}, true, 0, ui.Shift}},
	{"\033[<16;3;4M", MouseEvent{Pos{4, 3}, true, 0, ui.Ctrl}},
}

func TestReadEvent(t *testing.T) {
	r, w, cleanup := setupFileReader()
	defer cleanup()

	for _, test := range readEventTests {
		t.Run(test.input, func(t *testing.T) {
			w.WriteString(test.input)
			ev, err := readEvent(r)
			if ev != test.want {
				t.Errorf("got event %v, want %v", ev, test.want)
			}
			if err != nil {
				t.Errorf("got err %v, want %v", err, nil)
			}
		})
	}
}

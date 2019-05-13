package cliutil

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/tt"
)

var basicHandlerKeyEventsTests = []struct {
	name     string
	keys     []ui.Key
	wantCode string
	wantDot  int
}{
	{"ASCII characters",
		[]ui.Key{{Rune: 'a'}}, "a", 1},
	{"Unicode characters",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}},
		"代码", 6},
	{"Backspace",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Backspace}},
		"代", 3},
	{"Left 1",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}},
		"代码", 3},
	{"Left 2",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}, {Rune: ui.Left}},
		"代码", 0},
	{"Right",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}, {Rune: ui.Left},
			{Rune: ui.Right}},
		"代码", 3},
	{"Insert in the middle",
		[]ui.Key{{Rune: '['}, {Rune: ']'}, {Rune: ui.Left}, {Rune: 'x'}},
		"[x]", 2},
}

func TestBasicHandler_KeyEvents(t *testing.T) {
	for _, test := range basicHandlerKeyEventsTests {
		t.Run(test.name, func(t *testing.T) {
			st := clitypes.State{}
			for _, key := range test.keys {
				BasicHandler(term.KeyEvent(key), &st)
			}
			code, dot := st.CodeAndDot()
			if code != test.wantCode {
				t.Errorf("got code = %q, want %q", code, test.wantCode)
			}
			if dot != test.wantDot {
				t.Errorf("got dot = %v, want %v", dot, test.wantDot)
			}
		})
	}
}

func TestBasicHandler_NotifiesOnUnboundKeys(t *testing.T) {
	st := clitypes.State{}

	BasicHandler(term.KeyEvent{Mod: ui.Ctrl, Rune: 'X'}, &st)

	wantNotes := []string{"Unbound: Ctrl-X"}
	if notes := st.Raw.Notes; !reflect.DeepEqual(notes, wantNotes) {
		t.Errorf("Notes is %q, want %q", notes, wantNotes)
	}
}

var otherEvents = []term.Event{
	term.MouseEvent{}, term.RawRune('a'),
	term.PasteSetting(false), term.PasteSetting(true), term.CursorPosition{},
}

func TestBasicHandler_IgnoresOtherEvents(t *testing.T) {
	for _, event := range otherEvents {
		t.Run(fmt.Sprintf("event type %T", event), func(t *testing.T) {
			st := clitypes.State{}
			oldRaw := st.Raw
			BasicHandler(event, &st)
			if !reflect.DeepEqual(oldRaw, st.Raw) {
				t.Errorf("state mutated from %v to %v", oldRaw, st.Raw)
			}
		})
	}
}

func TestIsChar(t *testing.T) {
	tt.Test(t, tt.Fn("IsChar", IsChar), tt.Table{
		Args(ui.K(ui.Up)).Rets(false),
		Args(ui.K('A', ui.Ctrl)).Rets(false),
		Args(ui.K('a')).Rets(true),
		Args(ui.K('好')).Rets(true),
	})
}

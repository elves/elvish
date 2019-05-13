package insert

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
)

func TestModeLine_Default(t *testing.T) {
	m := &Mode{}
	if m.ModeLine() != nil {
		t.Errorf("got non-nil buf in default state")
	}
}

func TestModeLine_LiteralPaste(t *testing.T) {
	m := &Mode{}
	st := &clitypes.State{}

	m.HandleEvent(term.PasteSetting(true), st)
	if m.ModeLine() != literalPasteModeLine {
		t.Errorf("modeline != literalPasteModeLine during pasting")
	}

	m.HandleEvent(term.PasteSetting(false), st)
	if m.ModeLine() != nil {
		t.Errorf("modeline != nil after pasting")
	}
}

func TestModeLine_QuotePaste(t *testing.T) {
	m := &Mode{}
	m.Config.Raw.QuotePaste = true
	st := &clitypes.State{}

	m.HandleEvent(term.PasteSetting(true), st)

	if m.ModeLine() != quotePasteModeLine {
		t.Errorf("modeline != quotePasteModeLine during pasting with quote")
	}
}

func TestHandleEvent_LiteralPaste(t *testing.T) {
	testPaste(t, &Mode{}, "$100", "$100")
}

func TestHandleEvent_QuotePaste(t *testing.T) {
	m := &Mode{}
	m.Config.Raw.QuotePaste = true
	testPaste(t, m, "$100", parse.Quote("$100"))
}

func testPaste(t *testing.T, m *Mode, input, want string) {
	st := &clitypes.State{}
	st.Raw.Code = "[]"
	st.Raw.Dot = 1

	m.HandleEvent(term.PasteSetting(true), st)
	for _, r := range input {
		m.HandleEvent(term.KeyEvent{Rune: r}, st)
	}
	m.HandleEvent(term.PasteSetting(false), st)

	wantCode := "[" + want + "]"
	if st.Raw.Code != wantCode {
		t.Errorf("got code = %q, want %q", st.Raw.Code, wantCode)
	}
	if st.Raw.Dot != 1+len(want) {
		t.Errorf("got dot = %v, want %v", st.Raw.Dot, 1+len(want))
	}
}

var (
	events = []term.Event{
		term.KeyEvent{Rune: 'a'}, term.KeyEvent{Rune: 'b'}, term.MouseEvent{},
		term.KeyEvent{Rune: 'c'}, term.RawRune('d')}
	keyEvents = []ui.Key{{Rune: 'a'}, {Rune: 'b'}, {Rune: 'c'}}
)

func TestHandleEvent_CallsKeyHandler(t *testing.T) {
	var keys []ui.Key
	keyHandler := func(k ui.Key) clitypes.HandlerAction {
		keys = append(keys, k)
		return clitypes.NoAction
	}
	m := &Mode{KeyHandler: keyHandler}
	st := &clitypes.State{}

	for _, event := range events {
		m.HandleEvent(event, st)
	}
	if !reflect.DeepEqual(keys, keyEvents) {
		t.Errorf("got keys %v, want %v", keys, keyEvents)
	}
}

var abbrTests = []struct {
	name     string
	keys     []term.KeyEvent
	wantCode string
}{
	{"simple",
		[]term.KeyEvent{{Rune: 'x'}, {Rune: 'x'}},
		"> /dev/null"},
	{"suffix",
		[]term.KeyEvent{
			{Rune: 'l'}, {Rune: ' '}, {Rune: 'x'}, {Rune: 'x'}},
		"l > /dev/null"},
	{"longest suffix",
		[]term.KeyEvent{
			{Rune: 'l'}, {Rune: ' '}, {Rune: '2'}, {Rune: 'x'}, {Rune: 'x'}},
		"l > &2"},
	{"function key interrupts abbreviation",
		[]term.KeyEvent{
			{Rune: 'x'}, {Rune: ui.Left}, {Rune: ui.Right}, {Rune: 'x'}},
		"xx"},
}

func TestHandleEvent_ExpandsAbbr(t *testing.T) {
	m := &Mode{}
	m.AbbrIterate = func(f func(abbr, full string)) {
		f("xx", "> /dev/null")
		f("2xx", "> &2")
	}
	for _, test := range abbrTests {
		t.Run(test.name, func(t *testing.T) {
			st := &clitypes.State{}
			for _, k := range test.keys {
				m.HandleEvent(k, st)
			}
			if st.Raw.Code != test.wantCode {
				t.Errorf("got code = %q, want %q", st.Raw.Code, test.wantCode)
			}
		})
	}
}

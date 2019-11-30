package codearea

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

var bb = term.NewBufferBuilder

var renderTests = []el.RenderTest{
	{
		Name: "prompt only",
		Given: New(Spec{
			Prompt: ConstPrompt(ui.MakeText("~>", "bold"))}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("~>", "1").SetDotHere(),
	},
	{
		Name: "rprompt only",
		Given: New(Spec{
			RPrompt: ConstPrompt(ui.MakeText("RP", "inverse"))}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotHere().WriteSpaces(8).WriteStringSGR("RP", "7"),
	},
	{
		Name: "code only with dot at beginning",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "code", Dot: 0}}}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotHere().Write("code"),
	},
	{
		Name: "code only with dot at middle",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "code", Dot: 2}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("co").SetDotHere().Write("de"),
	},
	{
		Name: "code only with dot at end",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "prompt, code and rprompt",
		Given: New(Spec{
			Prompt:  ConstPrompt(ui.PlainText("~>")),
			RPrompt: ConstPrompt(ui.PlainText("RP")),
			State:   State{Buffer: Buffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere().Write("  RP"),
	},

	{
		Name: "prompt explicitly hidden ",
		Given: New(Spec{
			Prompt:  ConstPrompt(ui.PlainText("~>")),
			RPrompt: ConstPrompt(ui.PlainText("RP")),
			State:   State{Buffer: Buffer{Content: "code", Dot: 4}, HideRPrompt: true}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere(),
	},
	{
		Name: "rprompt too long",
		Given: New(Spec{
			Prompt:  ConstPrompt(ui.PlainText("~>")),
			RPrompt: ConstPrompt(ui.PlainText("1234")),
			State:   State{Buffer: Buffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere(),
	},
	{
		Name: "highlighted code",
		Given: New(Spec{
			Highlighter: func(code string) (ui.Text, []error) {
				return ui.MakeText(code, "bold"), nil
			},
			State: State{Buffer: Buffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("code", "1").SetDotHere(),
	},
	{
		Name: "static errors in code",
		Given: New(Spec{
			Prompt: ConstPrompt(ui.PlainText("> ")),
			Highlighter: func(code string) (ui.Text, []error) {
				err := errors.New("static error")
				return ui.PlainText(code), []error{err}
			},
			State: State{Buffer: Buffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("> code").SetDotHere().
			Newline().Write("static error"),
	},
	{
		Name: "pending code inserting at the dot",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 4},
			Pending: Pending{From: 4, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").WriteStringSGR("x", "4").SetDotHere(),
	},
	{
		Name: "pending code replacing at the dot",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 2},
			Pending: Pending{From: 2, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("co").WriteStringSGR("x", "4").SetDotHere(),
	},
	{
		Name: "pending code to the left of the dot",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 4},
			Pending: Pending{From: 1, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("c").WriteStringSGR("x", "4").Write("e").SetDotHere(),
	},
	{
		Name: "pending code to the right of the cursor",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 1},
			Pending: Pending{From: 2, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("c").SetDotHere().Write("o").
			WriteStringSGR("x", "4").Write("e"),
	},
	{
		Name: "ignore invalid pending code 1",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 4},
			Pending: Pending{From: 2, To: 1, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "ignore invalid pending code 2",
		Given: New(Spec{State: State{
			Buffer:  Buffer{Content: "code", Dot: 4},
			Pending: Pending{From: 5, To: 6, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "prioritize lines before the cursor with small height",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 2,
		Want: bb(10).Write("a").Newline().Write("b").SetDotHere(),
	},
	{
		Name: "show only the cursor line when height is 1",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 1,
		Want: bb(10).Write("b").SetDotHere(),
	},
	{
		Name: "show lines after the cursor when all lines before the cursor are shown",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 3,
		Want: bb(10).Write("a").Newline().Write("b").SetDotHere().
			Newline().Write("c"),
	},
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

var handleTests = []el.HandleTest{
	{
		Name:         "simple inserts",
		Given:        New(Spec{}),
		Events:       []term.Event{term.K('c'), term.K('o'), term.K('d'), term.K('e')},
		WantNewState: State{Buffer: Buffer{Content: "code", Dot: 4}},
	},
	{
		Name:         "unicode inserts",
		Given:        New(Spec{}),
		Events:       []term.Event{term.K('你'), term.K('好')},
		WantNewState: State{Buffer: Buffer{Content: "你好", Dot: 6}},
	},
	{
		Name:         "unterminated paste",
		Given:        New(Spec{}),
		Events:       []term.Event{term.PasteSetting(true), term.K('"'), term.K('x')},
		WantNewState: State{},
	},
	{
		Name:  "literal paste",
		Given: New(Spec{}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		WantNewState: State{Buffer: Buffer{Content: "\"x", Dot: 2}},
	},
	{
		Name:  "literal paste swallowing functional keys",
		Given: New(Spec{}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('a'), term.K(ui.F1), term.K('b'),
			term.PasteSetting(false)},
		WantNewState: State{Buffer: Buffer{Content: "ab", Dot: 2}},
	},
	{
		Name:  "quoted paste",
		Given: New(Spec{QuotePaste: func() bool { return true }}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		WantNewState: State{Buffer: Buffer{Content: "'\"x'", Dot: 4}},
	},
	{
		Name:  "backspace at end of code",
		Given: New(Spec{}),
		Events: []term.Event{
			term.K('c'), term.K('o'), term.K('d'), term.K('e'),
			term.K(ui.Backspace)},
		WantNewState: State{Buffer: Buffer{Content: "cod", Dot: 3}},
	},
	{
		Name: "backspace at middle of buffer",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "code", Dot: 2}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: State{Buffer: Buffer{Content: "cde", Dot: 1}},
	},
	{
		Name: "backspace at beginning of buffer",
		Given: New(Spec{State: State{
			Buffer: Buffer{Content: "code", Dot: 0}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: State{Buffer: Buffer{Content: "code", Dot: 0}},
	},
	{
		Name:  "backspace deleting unicode character",
		Given: New(Spec{}),
		Events: []term.Event{
			term.K('你'), term.K('好'), term.K(ui.Backspace)},
		WantNewState: State{Buffer: Buffer{Content: "你", Dot: 3}},
	},
	{
		Name: "abbreviation expansion",
		Given: New(Spec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K('n')},
		WantNewState: State{Buffer: Buffer{Content: "/dev/null", Dot: 9}},
	},
	{
		Name: "abbreviation expansion preferring longest",
		Given: New(Spec{
			Abbreviations: func(f func(abbr, full string)) {
				f("n", "none")
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K('n')},
		WantNewState: State{Buffer: Buffer{Content: "/dev/null", Dot: 9}},
	},
	{
		Name: "abbreviation expansion interrupted by function key",
		Given: New(Spec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K(ui.F1), term.K('n')},
		WantNewState: State{Buffer: Buffer{Content: "dn", Dot: 2}},
	},
	{
		Name: "overlay handler",
		Given: addOverlay(New(Spec{}), func(w *widget) el.Handler {
			return el.MapHandler{
				term.K('a'): func() { w.State.Buffer.InsertAtDot("b") },
			}
		}),
		Events:       []term.Event{term.K('a')},
		WantNewState: State{Buffer: Buffer{Content: "b", Dot: 1}},
	},
}

// A utility for building a Widget with an OverlayHandler as a single
// expression.
func addOverlay(w Widget, overlay func(*widget) el.Handler) Widget {
	ww := w.(*widget)
	ww.OverlayHandler = overlay(ww)
	return w
}

func TestHandle(t *testing.T) {
	el.TestHandle(t, handleTests)
}

var unhandledEvents = []term.Event{
	// Mouse events are unhandled
	term.MouseEvent{},
	// Function keys are unhandled (except Backspace)
	term.K(ui.F1),
	term.K('X', ui.Ctrl),
}

func TestHandle_UnhandledEvents(t *testing.T) {
	w := New(Spec{})
	for _, event := range unhandledEvents {
		handled := w.Handle(event)
		if handled {
			t.Errorf("event %v got handled", event)
		}
	}
}

func TestHandle_AbbreviationExpansionInterruptedByExternalMutation(t *testing.T) {
	w := New(Spec{
		Abbreviations: func(f func(abbr, full string)) {
			f("dn", "/dev/null")
		},
	})
	w.Handle(term.K('d'))
	w.MutateState(func(s *State) { s.Buffer.InsertAtDot("d") })
	w.Handle(term.K('n'))
	wantState := State{Buffer: Buffer{Content: "ddn", Dot: 3}}
	if state := w.CopyState(); !reflect.DeepEqual(state, wantState) {
		t.Errorf("got state %v, want %v", state, wantState)
	}
}

func TestHandle_EnterEmitsSubmit(t *testing.T) {
	submitted := false
	w := New(Spec{
		OnSubmit: func() { submitted = true },
		State:    State{Buffer: Buffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	if submitted != true {
		t.Errorf("OnSubmit not triggered")
	}
}

func TestHandle_DefaultNoopSubmit(t *testing.T) {
	w := New(Spec{State: State{
		Buffer: Buffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	// No panic, we are good
}

func TestState(t *testing.T) {
	w := New(Spec{})
	w.MutateState(func(s *State) { s.Buffer.Content = "code" })
	if w.CopyState().Buffer.Content != "code" {
		t.Errorf("state not mutated")
	}
}

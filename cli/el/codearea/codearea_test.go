package codearea

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderTests = []el.RenderTest{
	{
		Name: "prompt only",
		Given: New(Spec{
			Prompt: ConstPrompt(styled.MakeText("~>", "bold"))}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("~>", "1").SetDotToCursor(),
	},
	{
		Name: "rprompt only",
		Given: New(Spec{
			RPrompt: ConstPrompt(styled.MakeText("RP", "inverse"))}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotToCursor().WriteSpacesSGR(8, "").WriteStringSGR("RP", "7"),
	},
	{
		Name: "code only with dot at beginning",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 0}}}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotToCursor().WritePlain("code"),
	},
	{
		Name: "code only with dot at middle",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 2}}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("co").SetDotToCursor().WritePlain("de"),
	},
	{
		Name: "code only with dot at end",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor(),
	},
	{
		Name: "prompt, code and rprompt",
		Given: New(Spec{
			Prompt:  ConstPrompt(styled.Plain("~>")),
			RPrompt: ConstPrompt(styled.Plain("RP")),
			State:   State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("~>code").SetDotToCursor().WritePlain("  RP"),
	},
	{
		Name: "rprompt too long",
		Given: New(Spec{
			Prompt:  ConstPrompt(styled.Plain("~>")),
			RPrompt: ConstPrompt(styled.Plain("1234")),
			State:   State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("~>code").SetDotToCursor(),
	},
	{
		Name: "highlighted code",
		Given: New(Spec{
			Highlighter: func(code string) (styled.Text, []error) {
				return styled.MakeText(code, "bold"), nil
			},
			State: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("code", "1").SetDotToCursor(),
	},
	{
		Name: "static errors in code",
		Given: New(Spec{
			Prompt: ConstPrompt(styled.Plain("> ")),
			Highlighter: func(code string) (styled.Text, []error) {
				err := errors.New("static error")
				return styled.Plain(code), []error{err}
			},
			State: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("> code").SetDotToCursor().
			Newline().WritePlain("static error"),
	},
	{
		Name: "pending code inserting at the dot",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 4, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").WriteStringSGR("x", "4").SetDotToCursor(),
	},
	{
		Name: "pending code replacing at the dot",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 2},
			PendingCode: PendingCode{From: 2, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("co").WriteStringSGR("x", "4").SetDotToCursor(),
	},
	{
		Name: "pending code to the left of the dot",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 1, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("c").WriteStringSGR("x", "4").WritePlain("e").SetDotToCursor(),
	},
	{
		Name: "pending code to the right of the cursor",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 1},
			PendingCode: PendingCode{From: 2, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("c").SetDotToCursor().WritePlain("o").
			WriteStringSGR("x", "4").WritePlain("e"),
	},
	{
		Name: "ignore invalid pending code 1",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 2, To: 1, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor(),
	},
	{
		Name: "ignore invalid pending code 2",
		Given: New(Spec{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 5, To: 6, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor(),
	},
	{
		Name: "prioritize lines before the cursor with small height",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 2,
		Want: bb(10).WritePlain("a").Newline().WritePlain("b").SetDotToCursor(),
	},
	{
		Name: "show only the cursor line when height is 1",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 1,
		Want: bb(10).WritePlain("b").SetDotToCursor(),
	},
	{
		Name: "show lines after the cursor when all lines before the cursor are shown",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("a").Newline().WritePlain("b").SetDotToCursor().
			Newline().WritePlain("c"),
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
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}},
	},
	{
		Name:         "unicode inserts",
		Given:        New(Spec{}),
		Events:       []term.Event{term.K('你'), term.K('好')},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "你好", Dot: 6}},
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
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "\"x", Dot: 2}},
	},
	{
		Name:  "literal paste swallowing functional keys",
		Given: New(Spec{}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('a'), term.K(ui.F1), term.K('b'),
			term.PasteSetting(false)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "ab", Dot: 2}},
	},
	{
		Name:  "quoted paste",
		Given: New(Spec{QuotePaste: func() bool { return true }}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "'\"x'", Dot: 4}},
	},
	{
		Name:  "backspace at end of code",
		Given: New(Spec{}),
		Events: []term.Event{
			term.K('c'), term.K('o'), term.K('d'), term.K('e'),
			term.K(ui.Backspace)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "cod", Dot: 3}},
	},
	{
		Name: "backspace at middle of buffer",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 2}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "cde", Dot: 1}},
	},
	{
		Name: "backspace at beginning of buffer",
		Given: New(Spec{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 0}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 0}},
	},
	{
		Name:  "backspace deleting unicode character",
		Given: New(Spec{}),
		Events: []term.Event{
			term.K('你'), term.K('好'), term.K(ui.Backspace)},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "你", Dot: 3}},
	},
	{
		Name: "abbreviation expansion",
		Given: New(Spec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K('n')},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
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
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
	},
	{
		Name: "abbreviation expansion interrupted by function key",
		Given: New(Spec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K(ui.F1), term.K('n')},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "dn", Dot: 2}},
	},
	{
		Name: "overlay handler",
		Given: addOverlay(New(Spec{}), func(w *widget) el.Handler {
			return el.MapHandler{
				term.K('a'): func() { w.State.CodeBuffer.InsertAtDot("b") },
			}
		}),
		Events:       []term.Event{term.K('a')},
		WantNewState: State{CodeBuffer: CodeBuffer{Content: "b", Dot: 1}},
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
	w.MutateCodeAreaState(func(s *State) { s.CodeBuffer.InsertAtDot("d") })
	w.Handle(term.K('n'))
	wantState := State{CodeBuffer: CodeBuffer{Content: "ddn", Dot: 3}}
	if state := w.CopyState(); !reflect.DeepEqual(state, wantState) {
		t.Errorf("got state %v, want %v", state, wantState)
	}
}

func TestHandle_EnterEmitsSubmit(t *testing.T) {
	var submitted string
	w := New(Spec{
		OnSubmit: func(code string) { submitted = code },
		State:    State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	if submitted != "code" {
		t.Errorf("code not submitted")
	}
}

func TestHandle_DefaultNoopSubmit(t *testing.T) {
	w := New(Spec{State: State{
		CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	// No panic, we are good
}

func TestState(t *testing.T) {
	w := New(Spec{})
	w.MutateCodeAreaState(func(s *State) { s.CodeBuffer.Content = "code" })
	if w.CopyState().CodeBuffer.Content != "code" {
		t.Errorf("state not mutated")
	}
}

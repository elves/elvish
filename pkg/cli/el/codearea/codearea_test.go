package codearea

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

var bb = term.NewBufferBuilder

var renderTests = []el.RenderTest{
	{
		Name: "prompt only",
		Given: NewCodeArea(CodeAreaSpec{
			Prompt: ConstPrompt(ui.T("~>", ui.Bold))}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("~>", "1").SetDotHere(),
	},
	{
		Name: "rprompt only",
		Given: NewCodeArea(CodeAreaSpec{
			RPrompt: ConstPrompt(ui.T("RP", ui.Inverse))}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotHere().WriteSpaces(8).WriteStringSGR("RP", "7"),
	},
	{
		Name: "code only with dot at beginning",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "code", Dot: 0}}}),
		Width: 10, Height: 24,
		Want: bb(10).SetDotHere().Write("code"),
	},
	{
		Name: "code only with dot at middle",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "code", Dot: 2}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("co").SetDotHere().Write("de"),
	},
	{
		Name: "code only with dot at end",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "prompt, code and rprompt",
		Given: NewCodeArea(CodeAreaSpec{
			Prompt:  ConstPrompt(ui.T("~>")),
			RPrompt: ConstPrompt(ui.T("RP")),
			State:   CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere().Write("  RP"),
	},

	{
		Name: "prompt explicitly hidden ",
		Given: NewCodeArea(CodeAreaSpec{
			Prompt:  ConstPrompt(ui.T("~>")),
			RPrompt: ConstPrompt(ui.T("RP")),
			State:   CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}, HideRPrompt: true}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere(),
	},
	{
		Name: "rprompt too long",
		Given: NewCodeArea(CodeAreaSpec{
			Prompt:  ConstPrompt(ui.T("~>")),
			RPrompt: ConstPrompt(ui.T("1234")),
			State:   CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("~>code").SetDotHere(),
	},
	{
		Name: "highlighted code",
		Given: NewCodeArea(CodeAreaSpec{
			Highlighter: func(code string) (ui.Text, []error) {
				return ui.T(code, ui.Bold), nil
			},
			State: CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("code", "1").SetDotHere(),
	},
	{
		Name: "static errors in code",
		Given: NewCodeArea(CodeAreaSpec{
			Prompt: ConstPrompt(ui.T("> ")),
			Highlighter: func(code string) (ui.Text, []error) {
				err := errors.New("static error")
				return ui.T(code), []error{err}
			},
			State: CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("> code").SetDotHere().
			Newline().Write("static error"),
	},
	{
		Name: "pending code inserting at the dot",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 4},
			Pending: PendingCode{From: 4, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").WriteStringSGR("x", "4").SetDotHere(),
	},
	{
		Name: "pending code replacing at the dot",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 2},
			Pending: PendingCode{From: 2, To: 4, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("co").WriteStringSGR("x", "4").SetDotHere(),
	},
	{
		Name: "pending code to the left of the dot",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 4},
			Pending: PendingCode{From: 1, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("c").WriteStringSGR("x", "4").Write("e").SetDotHere(),
	},
	{
		Name: "pending code to the right of the cursor",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 1},
			Pending: PendingCode{From: 2, To: 3, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("c").SetDotHere().Write("o").
			WriteStringSGR("x", "4").Write("e"),
	},
	{
		Name: "ignore invalid pending code 1",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 4},
			Pending: PendingCode{From: 2, To: 1, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "ignore invalid pending code 2",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer:  CodeBuffer{Content: "code", Dot: 4},
			Pending: PendingCode{From: 5, To: 6, Content: "x"},
		}}),
		Width: 10, Height: 24,
		Want: bb(10).Write("code").SetDotHere(),
	},
	{
		Name: "prioritize lines before the cursor with small height",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 2,
		Want: bb(10).Write("a").Newline().Write("b").SetDotHere(),
	},
	{
		Name: "show only the cursor line when height is 1",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}}),
		Width: 10, Height: 1,
		Want: bb(10).Write("b").SetDotHere(),
	},
	{
		Name: "show lines after the cursor when all lines before the cursor are shown",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
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
		Given:        NewCodeArea(CodeAreaSpec{}),
		Events:       []term.Event{term.K('c'), term.K('o'), term.K('d'), term.K('e')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}},
	},
	{
		Name:         "unicode inserts",
		Given:        NewCodeArea(CodeAreaSpec{}),
		Events:       []term.Event{term.K('你'), term.K('好')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "你好", Dot: 6}},
	},
	{
		Name:         "unterminated paste",
		Given:        NewCodeArea(CodeAreaSpec{}),
		Events:       []term.Event{term.PasteSetting(true), term.K('"'), term.K('x')},
		WantNewState: CodeAreaState{},
	},
	{
		Name:  "literal paste",
		Given: NewCodeArea(CodeAreaSpec{}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "\"x", Dot: 2}},
	},
	{
		Name:  "literal paste swallowing functional keys",
		Given: NewCodeArea(CodeAreaSpec{}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('a'), term.K(ui.F1), term.K('b'),
			term.PasteSetting(false)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "ab", Dot: 2}},
	},
	{
		Name:  "quoted paste",
		Given: NewCodeArea(CodeAreaSpec{QuotePaste: func() bool { return true }}),
		Events: []term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "'\"x'", Dot: 4}},
	},
	{
		Name:  "backspace at end of code",
		Given: NewCodeArea(CodeAreaSpec{}),
		Events: []term.Event{
			term.K('c'), term.K('o'), term.K('d'), term.K('e'),
			term.K(ui.Backspace)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "cod", Dot: 3}},
	},
	{
		Name: "backspace at middle of buffer",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "code", Dot: 2}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "cde", Dot: 1}},
	},
	{
		Name: "backspace at beginning of buffer",
		Given: NewCodeArea(CodeAreaSpec{State: CodeAreaState{
			Buffer: CodeBuffer{Content: "code", Dot: 0}}}),
		Events:       []term.Event{term.K(ui.Backspace)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 0}},
	},
	{
		Name:  "backspace deleting unicode character",
		Given: NewCodeArea(CodeAreaSpec{}),
		Events: []term.Event{
			term.K('你'), term.K('好'), term.K(ui.Backspace)},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "你", Dot: 3}},
	},
	{
		Name: "abbreviation expansion",
		Given: NewCodeArea(CodeAreaSpec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K('n')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
	},
	{
		Name: "abbreviation expansion preferring longest",
		Given: NewCodeArea(CodeAreaSpec{
			Abbreviations: func(f func(abbr, full string)) {
				f("n", "none")
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K('n')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
	},
	{
		Name: "abbreviation expansion interrupted by function key",
		Given: NewCodeArea(CodeAreaSpec{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		}),
		Events:       []term.Event{term.K('d'), term.K(ui.F1), term.K('n')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "dn", Dot: 2}},
	},
	{
		Name: "overlay handler",
		Given: addOverlay(NewCodeArea(CodeAreaSpec{}), func(w *codeArea) el.Handler {
			return el.MapHandler{
				term.K('a'): func() { w.State.Buffer.InsertAtDot("b") },
			}
		}),
		Events:       []term.Event{term.K('a')},
		WantNewState: CodeAreaState{Buffer: CodeBuffer{Content: "b", Dot: 1}},
	},
}

// A utility for building a Widget with an OverlayHandler as a single
// expression.
func addOverlay(w CodeArea, overlay func(*codeArea) el.Handler) CodeArea {
	ww := w.(*codeArea)
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
	w := NewCodeArea(CodeAreaSpec{})
	for _, event := range unhandledEvents {
		handled := w.Handle(event)
		if handled {
			t.Errorf("event %v got handled", event)
		}
	}
}

func TestHandle_AbbreviationExpansionInterruptedByExternalMutation(t *testing.T) {
	w := NewCodeArea(CodeAreaSpec{
		Abbreviations: func(f func(abbr, full string)) {
			f("dn", "/dev/null")
		},
	})
	w.Handle(term.K('d'))
	w.MutateState(func(s *CodeAreaState) { s.Buffer.InsertAtDot("d") })
	w.Handle(term.K('n'))
	wantState := CodeAreaState{Buffer: CodeBuffer{Content: "ddn", Dot: 3}}
	if state := w.CopyState(); !reflect.DeepEqual(state, wantState) {
		t.Errorf("got state %v, want %v", state, wantState)
	}
}

func TestHandle_EnterEmitsSubmit(t *testing.T) {
	submitted := false
	w := NewCodeArea(CodeAreaSpec{
		OnSubmit: func() { submitted = true },
		State:    CodeAreaState{Buffer: CodeBuffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	if submitted != true {
		t.Errorf("OnSubmit not triggered")
	}
}

func TestHandle_DefaultNoopSubmit(t *testing.T) {
	w := NewCodeArea(CodeAreaSpec{State: CodeAreaState{
		Buffer: CodeBuffer{Content: "code", Dot: 4}}})
	w.Handle(term.K('\n'))
	// No panic, we are good
}

func TestState(t *testing.T) {
	w := NewCodeArea(CodeAreaSpec{})
	w.MutateState(func(s *CodeAreaState) { s.Buffer.Content = "code" })
	if w.CopyState().Buffer.Content != "code" {
		t.Errorf("state not mutated")
	}
}

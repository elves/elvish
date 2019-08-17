package codearea

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderTests = []clitypes.RenderTest{
	{
		Name: "prompt only",
		Given: &Widget{State: State{
			Prompt: styled.MakeText("~>", "bold")}},
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("~>", "1").SetDotToCursor(),
	},
	{
		Name: "rprompt only",
		Given: &Widget{State: State{
			RPrompt: styled.MakeText("RP", "inverse")}},
		Width: 10, Height: 24,
		Want: bb(10).SetDotToCursor().WriteSpacesSGR(8, "").WriteStringSGR("RP", "7"),
	},
	{
		Name: "code only with dot at beginning",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 0}}},
		Width: 10, Height: 24,
		Want: bb(10).SetDotToCursor().WritePlain("code"),
	},
	{
		Name: "code only with dot at middle",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 2}}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("co").SetDotToCursor().WritePlain("de"),
	},
	{
		Name: "code only with dot at end",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor(),
	},
	{
		Name: "prompt, code and rprompt",
		Given: &Widget{State: State{
			Prompt:     styled.Plain("~>"),
			CodeBuffer: CodeBuffer{Content: "code", Dot: 4},
			RPrompt:    styled.Plain("RP"),
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("~>code").SetDotToCursor().WritePlain("  RP"),
	},
	{
		Name: "rprompt too long",
		Given: &Widget{State: State{
			Prompt:     styled.Plain("~>"),
			CodeBuffer: CodeBuffer{Content: "code", Dot: 4},
			RPrompt:    styled.Plain("1234"),
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("~>code").SetDotToCursor(),
	},
	{
		Name: "highlighted code",
		Given: &Widget{
			State: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}},
			Highlighter: func(code string) (styled.Text, []error) {
				return styled.MakeText(code, "bold"), nil
			},
		},
		Width: 10, Height: 24,
		Want: bb(10).WriteStringSGR("code", "1").SetDotToCursor(),
	},
	{
		Name: "static errors in code",
		Given: &Widget{
			State: State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}},
			Highlighter: func(code string) (styled.Text, []error) {
				err := errors.New("static error")
				return styled.Plain(code), []error{err}
			},
		},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor().
			Newline().WritePlain("static error"),
	},
	{
		Name: "pending code inserting at the dot",
		Given: &Widget{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 4, To: 4, Content: "x"},
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").WriteStringSGR("x", "4").SetDotToCursor(),
	},
	{
		Name: "pending code replacing at the dot",
		Given: &Widget{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 2},
			PendingCode: PendingCode{From: 2, To: 4, Content: "x"},
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("co").WriteStringSGR("x", "4").SetDotToCursor(),
	},
	{
		Name: "pending code to the left of the dot",
		Given: &Widget{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 1, To: 3, Content: "x"},
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("c").WriteStringSGR("x", "4").WritePlain("e").SetDotToCursor(),
	},
	{
		Name: "pending code to the right of the cursor",
		Given: &Widget{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 1},
			PendingCode: PendingCode{From: 2, To: 3, Content: "x"},
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("c").SetDotToCursor().WritePlain("o").
			WriteStringSGR("x", "4").WritePlain("e"),
	},
	{
		Name: "ignore invalid pending code",
		Given: &Widget{State: State{
			CodeBuffer:  CodeBuffer{Content: "code", Dot: 4},
			PendingCode: PendingCode{From: 2, To: 1, Content: "x"},
		}},
		Width: 10, Height: 24,
		Want: bb(10).WritePlain("code").SetDotToCursor(),
	},
	{
		Name: "prioritize lines before the cursor with small height",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}},
		Width: 10, Height: 2,
		Want: bb(10).WritePlain("a").Newline().WritePlain("b").SetDotToCursor(),
	},
	{
		Name: "show only the cursor line when height is 1",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}},
		Width: 10, Height: 1,
		Want: bb(10).WritePlain("b").SetDotToCursor(),
	},
	{
		Name: "show lines after the cursor when all lines before the cursor are shown",
		Given: &Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "a\nb\nc\nd", Dot: 3},
		}},
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("a").Newline().WritePlain("b").SetDotToCursor().
			Newline().WritePlain("c"),
	},
}

func TestRender(t *testing.T) {
	clitypes.TestRender(t, renderTests)
}

var handleTests = []struct {
	name      string
	widget    *Widget
	events    []term.Event
	wantState State
}{
	{
		"simple inserts",
		&Widget{},
		[]term.Event{term.K('c'), term.K('o'), term.K('d'), term.K('e')},
		State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}},
	},
	{
		"unicode inserts",
		&Widget{},
		[]term.Event{term.K('你'), term.K('好')},
		State{CodeBuffer: CodeBuffer{Content: "你好", Dot: 6}},
	},
	{
		"unterminated paste",
		&Widget{},
		[]term.Event{term.PasteSetting(true), term.K('"'), term.K('x')},
		State{},
	},
	{
		"literal paste",
		&Widget{},
		[]term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		State{CodeBuffer: CodeBuffer{Content: "\"x", Dot: 2}},
	},
	{
		"literal paste swallowing functional keys",
		&Widget{},
		[]term.Event{
			term.PasteSetting(true),
			term.K('a'), term.K(ui.F1), term.K('b'),
			term.PasteSetting(false)},
		State{CodeBuffer: CodeBuffer{Content: "ab", Dot: 2}},
	},
	{
		"quoted paste",
		&Widget{QuotePaste: func() bool { return true }},
		[]term.Event{
			term.PasteSetting(true),
			term.K('"'), term.K('x'),
			term.PasteSetting(false)},
		State{CodeBuffer: CodeBuffer{Content: "'\"x'", Dot: 4}},
	},
	{
		"backspace at end of code",
		&Widget{},
		[]term.Event{
			term.K('c'), term.K('o'), term.K('d'), term.K('e'),
			term.K(ui.Backspace)},
		State{CodeBuffer: CodeBuffer{Content: "cod", Dot: 3}},
	},
	{
		"backspace at middle of buffer",
		&Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 2}}},
		[]term.Event{term.K(ui.Backspace)},
		State{CodeBuffer: CodeBuffer{Content: "cde", Dot: 1}},
	},
	{
		"backspace at beginning of buffer",
		&Widget{State: State{
			CodeBuffer: CodeBuffer{Content: "code", Dot: 0}}},
		[]term.Event{term.K(ui.Backspace)},
		State{CodeBuffer: CodeBuffer{Content: "code", Dot: 0}},
	},
	{
		"backspace deleting unicode character",
		&Widget{},
		[]term.Event{
			term.K('你'), term.K('好'), term.K(ui.Backspace)},
		State{CodeBuffer: CodeBuffer{Content: "你", Dot: 3}},
	},
	{
		"abbreviation expansion",
		&Widget{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		},
		[]term.Event{term.K('d'), term.K('n')},
		State{CodeBuffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
	},
	{
		"abbreviation expansion preferring longest",
		&Widget{
			Abbreviations: func(f func(abbr, full string)) {
				f("n", "none")
				f("dn", "/dev/null")
			},
		},
		[]term.Event{term.K('d'), term.K('n')},
		State{CodeBuffer: CodeBuffer{Content: "/dev/null", Dot: 9}},
	},
	{
		"abbreviation expansion interrupted by function key",
		&Widget{
			Abbreviations: func(f func(abbr, full string)) {
				f("dn", "/dev/null")
			},
		},
		[]term.Event{term.K('d'), term.K(ui.F1), term.K('n')},
		State{CodeBuffer: CodeBuffer{Content: "dn", Dot: 2}},
	},
}

func TestHandle(t *testing.T) {
	for _, test := range handleTests {
		t.Run(test.name, func(t *testing.T) {
			w := test.widget
			for _, event := range test.events {
				w.Handle(event)
			}
			if !reflect.DeepEqual(w.State, test.wantState) {
				t.Errorf("got state %v, want %v", w.State, test.wantState)
			}
		})
	}
}

var unhandledEvents = []term.Event{
	// Mouse events are unhandled
	term.MouseEvent{},
	// Function keys are unhandled (except Backspace)
	term.K(ui.F1),
	term.K('X', ui.Ctrl),
}

func TestHandle_UnhandledEvents(t *testing.T) {
	w := &Widget{}
	for _, event := range unhandledEvents {
		handled := w.Handle(event)
		if handled {
			t.Errorf("event %v got handled", event)
		}
	}
}

func TestHandle_AbbreviationExpansionInterruptedByExternalMutation(t *testing.T) {
	w := &Widget{
		Abbreviations: func(f func(abbr, full string)) {
			f("dn", "/dev/null")
		},
	}
	w.Handle(term.K('d'))
	w.MutateCodeAreaState(func(s *State) { s.CodeBuffer.InsertAtDot("d") })
	w.Handle(term.K('n'))
	wantState := State{CodeBuffer: CodeBuffer{Content: "ddn", Dot: 3}}
	if !reflect.DeepEqual(w.State, wantState) {
		t.Errorf("got state %v, want %v", w.State, wantState)
	}
}

func TestHandle_EnterEmitsSubmit(t *testing.T) {
	var submitted string
	w := &Widget{
		State:    State{CodeBuffer: CodeBuffer{Content: "code", Dot: 4}},
		OnSubmit: func(code string) { submitted = code },
	}
	w.Handle(term.K('\n'))
	if submitted != "code" {
		t.Errorf("code not submitted")
	}
}

func TestHandle_DefaultNoopSubmit(t *testing.T) {
	w := &Widget{State: State{
		CodeBuffer: CodeBuffer{Content: "code", Dot: 4}}}
	w.Handle(term.K('\n'))
	// No panic, we are good
}

func TestState(t *testing.T) {
	w := &Widget{}
	w.MutateCodeAreaState(func(s *State) { s.CodeBuffer.Content = "code" })
	if w.CopyState().CodeBuffer.Content != "code" {
		t.Errorf("state not mutated")
	}
}

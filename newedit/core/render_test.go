package core

import (
	"errors"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestRenderers(t *testing.T) {
	tt.Test(t, tt.Fn("Render", ui.Render), tt.Table{
		Args(&notesRenderer{[]string{
			"note 1", "long note 2",
		}}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteString("note 1\n", "").
				WriteString("long note 2", "").
				Buffer()),

		// TODO: mainRenderer

		// codeContentRenderer: Prompt and code, with indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: "> "}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				SetIndent(2).
				SetEagerWrap(true).
				WriteString("> abcdefg", "").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Multi-line prompt and code, without indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: ">\n"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				// No indent as the prompt is multi-line
				SetEagerWrap(true).
				WriteString(">\n", "").
				WriteString("abcdefg", "").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Long prompt and code, without indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: ">>> "}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				// No indent as the prompt is too long
				SetEagerWrap(true).
				WriteString(">>> abcdefg", "").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Visible rprompt
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abc"}}, dot: 3,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteString("abc", "").
				SetDotToCursor().
				WriteString("  RP", "").
				Buffer()),

		// codeContentRenderer: Rprompt hidden as no padding available (negative
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdef"}}, dot: 6,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteString("abcdef", "").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Rprompt hidden as no padding available (zero
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcde"}}, dot: 5,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteString("abcde", "").
				SetDotToCursor().
				Buffer()),

		Args(&codeErrorsRenderer{[]error{
			errors.New("error 1"),
			errors.New("long error 2"),
		}}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteString("error 1\n", "").
				WriteString("long error 2", "").
				Buffer()),
	})
}

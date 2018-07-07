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
		// mainRenderer: No modeline, no listing, enough height - result is the
		// same as bufCode
		Args(&mainRenderer{
			maxHeight: 10,
			bufCode: ui.NewBufferBuilder(7).
				WriteUnstyled("some code").SetDotToCursor().Buffer(),
			mode: &fakeMode{},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("some code").SetDotToCursor().Buffer()),

		// mainRenderer: No modeline, no listing, not enough height - show
		// lines close to where the dot is on
		Args(&mainRenderer{
			maxHeight: 2,
			bufCode: ui.NewBufferBuilder(7).
				WriteUnstyled("line 1").Newline().
				WriteUnstyled("line 2").Newline().
				WriteUnstyled("line 3").SetDotToCursor().Buffer(),
			mode: &fakeMode{},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("line 2").Newline().
				WriteUnstyled("line 3").SetDotToCursor().Buffer()),

		// mainRenderer: Modeline, no listing, enough height - result is the
		// bufCode + bufMode
		Args(&mainRenderer{
			maxHeight: 10,
			bufCode: ui.NewBufferBuilder(7).
				WriteUnstyled("some code").SetDotToCursor().Buffer(),
			mode: &fakeMode{
				modeLine: &linesRenderer{[]string{"MODE"}},
			},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("some code").SetDotToCursor().Newline().
				WriteUnstyled("MODE").Buffer()),

		// codeContentRenderer: Prompt and code, with indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: "> "}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				SetIndent(2).
				SetEagerWrap(true).
				WriteUnstyled("> abcdefg").
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
				WriteUnstyled(">\n").
				WriteUnstyled("abcdefg").
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
				WriteUnstyled(">>> abcdefg").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Visible rprompt
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abc"}}, dot: 3,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("abc").
				SetDotToCursor().
				WriteUnstyled("  RP").
				Buffer()),

		// codeContentRenderer: Rprompt hidden as no padding available (negative
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdef"}}, dot: 6,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("abcdef").
				SetDotToCursor().
				Buffer()),

		// codeContentRenderer: Rprompt hidden as no padding available (zero
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcde"}}, dot: 5,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("abcde").
				SetDotToCursor().
				Buffer()),

		Args(&linesRenderer{[]string{
			"note 1", "long note 2",
		}}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("note 1\n").
				WriteUnstyled("long note 2").
				Buffer()),

		Args(&codeErrorsRenderer{[]error{
			errors.New("error 1"),
			errors.New("long error 2"),
		}}, 7).
			Rets(ui.NewBufferBuilder(7).
				WriteUnstyled("error 1\n").
				WriteUnstyled("long error 2").
				Buffer()),
	})
}

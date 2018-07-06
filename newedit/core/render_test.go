package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

var wantBuf0, wantBuf20, wantBuf21, wantBuf22, wantBuf23, wantBuf24, wantBuf25, wantBuf3 *ui.Buffer

func TestRenderers(t *testing.T) {
	tt.Test(t, tt.Fn("Render", ui.Render), tt.Table{
		Args(&notesRenderer{[]string{
			"note 1", "long note 2",
		}}, 7).Rets(matchBufferLines{wantBuf0}),

		// TODO: mainRenderer

		// codeContentRenderer: Prompt and code, with indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: "> "}},
		}, 7).Rets(matchBuffer{wantBuf20}),

		// codeContentRenderer: Multi-line prompt and code, without indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: ">\n"}},
		}, 7).Rets(matchBuffer{wantBuf21}),

		// codeContentRenderer: Long prompt and code, without indentation
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdefg"}}, dot: 7,
			prompt: styled.Text{styled.Segment{Text: ">>> "}},
		}, 7).Rets(matchBuffer{wantBuf22}),

		// codeContentRenderer: Visible rprompt
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abc"}}, dot: 3,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).Rets(matchBuffer{wantBuf23}),

		// codeContentRenderer: Rprompt hidden as no padding available (negative
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcdef"}}, dot: 6,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).Rets(matchBuffer{wantBuf24}),

		// codeContentRenderer: Rprompt hidden as no padding available (zero
		// padding)
		Args(&codeContentRenderer{
			code: styled.Text{styled.Segment{Text: "abcde"}}, dot: 5,
			rprompt: styled.Text{styled.Segment{Text: "RP"}},
		}, 7).Rets(matchBuffer{wantBuf25}),

		Args(&codeErrorsRenderer{[]error{
			errors.New("error 1"),
			errors.New("long error 2"),
		}}, 7).Rets(matchBufferLines{wantBuf3}),
	})
}

func init() {
	wantBuf0 = ui.NewBufferBuilder(7).
		WriteString("note 1\n", "").
		WriteString("long note 2", "").
		Buffer()

	wantBuf20 = ui.NewBufferBuilder(7).
		SetIndent(2).
		SetEagerWrap(true).
		WriteString("> abcdefg", "").
		SetDotToCursor().
		Buffer()

	wantBuf21 = ui.NewBufferBuilder(7).
		// No indent as the prompt is multi-line
		SetEagerWrap(true).
		WriteString(">\n", "").
		WriteString("abcdefg", "").
		SetDotToCursor().
		Buffer()

	wantBuf22 = ui.NewBufferBuilder(7).
		// No indent as the prompt is too long
		SetEagerWrap(true).
		WriteString(">>> abcdefg", "").
		SetDotToCursor().
		Buffer()

	wantBuf23 = ui.NewBufferBuilder(7).
		WriteString("abc", "").
		SetDotToCursor().
		WriteString("  RP", "").
		Buffer()

	wantBuf24 = ui.NewBufferBuilder(7).
		WriteString("abcdef", "").
		SetDotToCursor().
		Buffer()

	wantBuf25 = ui.NewBufferBuilder(7).
		WriteString("abcde", "").
		SetDotToCursor().
		Buffer()

	wantBuf3 = ui.NewBufferBuilder(7).
		WriteString("error 1\n", "").
		WriteString("long error 2", "").
		Buffer()
}

type matchBuffer struct {
	want *ui.Buffer
}

func (m matchBuffer) Match(v tt.RetValue) bool {
	buf := v.(*ui.Buffer)
	return buf.Dot == m.want.Dot && reflect.DeepEqual(buf.Lines, m.want.Lines)
}

type matchBufferLines struct {
	want *ui.Buffer
}

func (m matchBufferLines) Match(v tt.RetValue) bool {
	buf := v.(*ui.Buffer)
	return reflect.DeepEqual(buf.Lines, m.want.Lines)
}

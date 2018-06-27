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
	wantBuf0 = ui.NewBuffer(7)
	wantBuf0.WriteString("note 1\n", "")
	wantBuf0.WriteString("long note 2", "")

	wantBuf20 = ui.NewBuffer(7)
	wantBuf20.Indent = 2
	wantBuf20.EagerWrap = true
	wantBuf20.WriteString("> abcdefg", "")
	wantBuf20.SetDot(wantBuf20.Cursor())

	wantBuf21 = ui.NewBuffer(7)
	// No indent as the prompt is multi-line
	wantBuf21.EagerWrap = true
	wantBuf21.WriteString(">\n", "")
	wantBuf21.WriteString("abcdefg", "")
	wantBuf21.SetDot(wantBuf21.Cursor())

	wantBuf22 = ui.NewBuffer(7)
	// No indent as the prompt is too long
	wantBuf22.EagerWrap = true
	wantBuf22.WriteString(">>> abcdefg", "")
	wantBuf22.SetDot(wantBuf22.Cursor())

	wantBuf23 = ui.NewBuffer(7)
	wantBuf23.WriteString("abc", "")
	wantBuf23.SetDot(wantBuf23.Cursor())
	wantBuf23.WriteString("  RP", "")

	wantBuf24 = ui.NewBuffer(7)
	wantBuf24.WriteString("abcdef", "")
	wantBuf24.SetDot(wantBuf24.Cursor())

	wantBuf25 = ui.NewBuffer(7)
	wantBuf25.WriteString("abcde", "")
	wantBuf25.SetDot(wantBuf25.Cursor())

	wantBuf3 = ui.NewBuffer(7)
	wantBuf3.WriteString("error 1\n", "")
	wantBuf3.WriteString("long error 2", "")
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

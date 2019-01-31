package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

type dummyRenderer struct{}

func (dummyRenderer) Render(bb *BufferBuilder) { bb.WriteString("dummy", "1") }

var Args = tt.Args

func TestRender(t *testing.T) {
	tt.Test(t, tt.Fn("Render", Render), tt.Table{
		Args(dummyRenderer{}, 10).
			Rets(NewBufferBuilder(10).WriteString("dummy", "1").Buffer()),

		Args(NewStringRenderer("string"), 10).
			Rets(NewBufferBuilder(10).WriteString("string", "").Buffer()),
		Args(NewStringRenderer("string"), 3).
			Rets(NewBufferBuilder(3).WriteString("str", "").Buffer()),

		Args(NewLinesRenderer("line 1", "line 2"), 10).
			Rets(
				NewBufferBuilder(10).WriteString("line 1", "").Newline().
					WriteString("line 2", "").Buffer()),
		Args(NewLinesRenderer("line 1", "line 2"), 3).
			Rets(
				NewBufferBuilder(3).WriteString("lin", "").Newline().
					WriteString("lin", "").Buffer()),

		Args(NewModeLineRenderer("M", "f"), 10).
			Rets(
				NewBufferBuilder(10).
					WriteString("M", styleForMode.String()).
					WriteSpaces(1, "").
					WriteString("f", styleForFilter.String()).
					SetDotToCursor().
					Buffer()),

		// Width left for scrollbar is 5
		Args(NewModeLineWithScrollBarRenderer(NewModeLineRenderer("M", "f"), 5, 0, 1), 10).
			Rets(
				NewBufferBuilder(10).
					WriteString("M", styleForMode.String()).
					WriteSpaces(1, "").
					WriteString("f", styleForFilter.String()).
					SetDotToCursor().
					WriteSpaces(1, "").
					Write(' ', styleForScrollBarThumb.String()).
					WriteString("━━━━", styleForScrollBarArea.String()).
					Buffer()),

		Args(NewRendererWithVerticalScrollbar(NewLinesRenderer("1", "2", "3"), 3, 0, 1), 5).
			Rets(
				NewBufferBuilder(5).
					WriteString("1   ", "").
					Write(' ', styleForScrollBarThumb.String()).
					Newline().
					WriteString("2   ", "").
					Write('│', styleForScrollBarArea.String()).
					Newline().
					WriteString("3   ", "").
					Write('│', styleForScrollBarArea.String()).
					Buffer()),
	})
}

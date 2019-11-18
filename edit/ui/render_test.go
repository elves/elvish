package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

type dummyRenderer struct{}

func (dummyRenderer) Render(bb *BufferBuilder) { bb.WriteStringSGR("dummy", "1") }

var Args = tt.Args

func TestRender(t *testing.T) {
	tt.Test(t, tt.Fn("Render", Render), tt.Table{
		Args(dummyRenderer{}, 10).
			Rets(NewBufferBuilder(10).WriteStringSGR("dummy", "1").Buffer()),

		Args(NewStringRenderer("string"), 10).
			Rets(NewBufferBuilder(10).WriteStringSGR("string", "").Buffer()),
		Args(NewStringRenderer("string"), 3).
			Rets(NewBufferBuilder(3).WriteStringSGR("str", "").Buffer()),

		Args(NewLinesRenderer("line 1", "line 2"), 10).
			Rets(
				NewBufferBuilder(10).WriteStringSGR("line 1", "").Newline().
					WriteStringSGR("line 2", "").Buffer()),
		Args(NewLinesRenderer("line 1", "line 2"), 3).
			Rets(
				NewBufferBuilder(3).WriteStringSGR("lin", "").Newline().
					WriteStringSGR("lin", "").Buffer()),

		Args(NewModeLineRenderer("M", "f"), 10).
			Rets(
				NewBufferBuilder(10).
					WriteStringSGR("M", styleForMode.String()).
					WriteSpacesSGR(1, "").
					WriteStringSGR("f", styleForFilter.String()).
					SetDotHere().
					Buffer()),

		// Width left for scrollbar is 5
		Args(NewModeLineWithScrollBarRenderer(NewModeLineRenderer("M", "f"), 5, 0, 1), 10).
			Rets(
				NewBufferBuilder(10).
					WriteStringSGR("M", styleForMode.String()).
					WriteSpacesSGR(1, "").
					WriteStringSGR("f", styleForFilter.String()).
					SetDotHere().
					WriteSpacesSGR(1, "").
					WriteRuneSGR(' ', styleForScrollBarThumb.String()).
					WriteStringSGR("━━━━", styleForScrollBarArea.String()).
					Buffer()),

		Args(NewRendererWithVerticalScrollbar(NewLinesRenderer("1", "2", "3"), 3, 0, 1), 5).
			Rets(
				NewBufferBuilder(5).
					WriteStringSGR("1   ", "").
					WriteRuneSGR(' ', styleForScrollBarThumb.String()).
					Newline().
					WriteStringSGR("2   ", "").
					WriteRuneSGR('│', styleForScrollBarArea.String()).
					Newline().
					WriteStringSGR("3   ", "").
					WriteRuneSGR('│', styleForScrollBarArea.String()).
					Buffer()),
	})
}

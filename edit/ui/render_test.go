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
	})
}

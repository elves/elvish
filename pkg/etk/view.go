package etk

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// View is the appearance of a component.
type View interface {
	// Render renders onto a region of bound width and height.
	Render(width, height int) *term.Buffer
}

func EmptyView() View { return emptyView{} }

type emptyView struct{}

func (e emptyView) Render(width, height int) *term.Buffer { return term.NewBuffer(width) }

func TextView(dotBefore int, spans ...ui.Text) View { return textView{spans, dotBefore} }

type textView struct {
	Spans     []ui.Text
	DotBefore int
}

func (t textView) Render(width, height int) *term.Buffer {
	bb := term.NewBufferBuilder(width)
	for i, span := range t.Spans {
		bb.WriteStyled(span)
		if i+1 == t.DotBefore {
			bb.SetDotHere()
		}
	}
	buf := bb.Buffer()
	// TODO: dot line
	buf.TrimToLines(0, height)
	return buf
}

func VBoxView(focus int, rows ...View) View { return vboxView{rows, focus} }

type vboxView struct {
	Rows  []View
	Focus int
}

func (v vboxView) Render(width, height int) *term.Buffer {
	if len(v.Rows) == 0 {
		return term.NewBuffer(width)
	}
	buf := v.Rows[0].Render(width, height-len(v.Rows)-1)
	for i := 1; i < len(v.Rows); i++ {
		rowHeight := height - len(buf.Lines) - len(v.Rows) - i - 1
		if rowHeight <= 0 {
			break
		}
		buf.Extend(v.Rows[i].Render(width, rowHeight), i == v.Focus)
	}
	return buf
}

func HBoxView(focus int, cols ...View) View { return hboxView{cols, focus} }

type hboxView struct {
	Cols  []View
	Focus int
}

func (h hboxView) Render(width, height int) *term.Buffer {
	if len(h.Cols) == 0 {
		return term.NewBuffer(width)
	}
	colWidth := width / len(h.Cols)

	buf := h.Cols[0].Render(colWidth, height)
	for i := 1; i < len(h.Cols); i++ {
		// TODO: Focus
		buf.ExtendRight(h.Cols[i].Render(colWidth, height))
	}
	return buf
}

func HBoxFlexView(focus, gap int, cols ...View) View { return hboxFlexView{cols, focus, gap} }

type hboxFlexView struct {
	Cols  []View
	Focus int
	Gap   int
}

func (h hboxFlexView) Render(width, height int) *term.Buffer {
	buf := term.NewBuffer(0)
	if len(h.Cols) == 0 {
		return buf
	}
	// TODO: Handle very narrow width

	for i, col := range h.Cols {
		bufCol := col.Render(width-(h.Gap+1)*(len(h.Cols)-i-1), height)
		actualWidth := term.CellsWidth(bufCol.Lines[0])
		for _, line := range bufCol.Lines[1:] {
			actualWidth = max(actualWidth, term.CellsWidth(line))
		}
		bufCol.Width = actualWidth
		// TODO: Focus
		if i > 0 {
			buf.Width += h.Gap
		}
		buf.ExtendRight(bufCol)
	}
	return buf
}

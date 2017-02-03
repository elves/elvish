package edit

type renderer interface {
	render(b *buffer)
}

func render(r renderer, width int) *buffer {
	if r == nil {
		return nil
	}
	b := newBuffer(width)
	r.render(b)
	return b
}

type modeLine struct {
	title  string
	filter string
}

func (ml modeLine) render(b *buffer) {
	b.writes(ml.title, styleForMode.String())
	b.writes(" ", "")
	b.writes(ml.filter, styleForFilter.String())
	b.dot = b.cursor()
}

type modeLineWithScrollBar struct {
	modeLine
	n, low, high int
}

func (ml modeLineWithScrollBar) render(b *buffer) {
	ml.modeLine.render(b)

	scrollbarWidth := b.width - lineWidth(b.cells[len(b.cells)-1]) - 2
	if scrollbarWidth >= 3 {
		b.writes(" ", "")
		writeHorizontalScrollbar(b, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}

package edit

import (
	"container/list"

	"github.com/elves/elvish/util"
)

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

type modeLineRenderer struct {
	title  string
	filter string
}

func (ml modeLineRenderer) render(b *buffer) {
	b.writes(ml.title, styleForMode.String())
	b.writes(" ", "")
	b.writes(ml.filter, styleForFilter.String())
	b.dot = b.cursor()
}

type modeLineWithScrollBarRenderer struct {
	modeLineRenderer
	n, low, high int
}

func (ml modeLineWithScrollBarRenderer) render(b *buffer) {
	ml.modeLineRenderer.render(b)

	scrollbarWidth := b.width - lineWidth(b.cells[len(b.cells)-1]) - 2
	if scrollbarWidth >= 3 {
		b.writes(" ", "")
		writeHorizontalScrollbar(b, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}

type placeholderRenderer string

func (lp placeholderRenderer) render(b *buffer) {
	b.writes(util.TrimWcwidth(string(lp), b.width), "")
}

type listingRenderer struct {
	// A List of styled items.
	list.List
}

func (ls listingRenderer) render(b *buffer) {
	for p := ls.Front(); p != nil; p = p.Next() {
		line := p.Value.(styled)
		if p != ls.Front() {
			b.newline()
		}
		b.writes(util.ForceWcwidth(line.text, b.width), line.styles.String())
	}
}

type listingWithScrollBarRenderer struct {
	listingRenderer
	n, low, high, height int
}

func (ls listingWithScrollBarRenderer) render(b *buffer) {
	b1 := render(ls.listingRenderer, b.width-1)
	b.extendHorizontal(b1, 0)

	scrollbar := renderScrollbar(ls.n, ls.low, ls.high, ls.height)
	b.extendHorizontal(scrollbar, b.width-1)
}

type navRenderer struct {
	maxHeight                      int
	fwParent, fwCurrent, fwPreview int
	parent, current, preview       renderer
}

func makeNavRenderer(h int, w1, w2, w3 int, r1, r2, r3 renderer) renderer {
	return &navRenderer{h, w1, w2, w3, r1, r2, r3}
}

func (nr *navRenderer) render(b *buffer) {
	margin := navigationListingColMargin

	w := b.width - margin*2
	ws := distributeWidths(w,
		[]float64{parentColumnWeight, currentColumnWeight, previewColumnWeight},
		[]int{nr.fwParent, nr.fwCurrent, nr.fwPreview},
	)
	wParent, wCurrent, wPreview := ws[0], ws[1], ws[2]

	bParent := render(nr.parent, wParent)
	b.extendHorizontal(bParent, 0)

	bCurrent := render(nr.current, wCurrent)
	b.extendHorizontal(bCurrent, wParent+margin)

	if wPreview > 0 {
		bPreview := render(nr.preview, wPreview)
		b.extendHorizontal(bPreview, wParent+wCurrent+2*margin)
	}
}

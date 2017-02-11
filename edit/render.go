package edit

import (
	"container/list"
	"strings"
	"unicode/utf8"

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

// editorRenderer renders the entire editor.
type editorRenderer struct {
	*editorState
	height  int
	bufNoti *buffer
}

func (er *editorRenderer) render(buf *buffer) {
	height, width, es := er.height, buf.width, er.editorState

	mode := es.mode.Mode()

	var bufNoti, bufLine, bufMode, bufTips, bufListing *buffer
	// butNoti
	if len(es.notifications) > 0 {
		bufNoti = newBuffer(width)
		bufNoti.writes(strings.Join(es.notifications, "\n"), "")
		es.notifications = nil
	}

	// bufLine
	b := newBuffer(width)
	bufLine = b

	b.newlineWhenFull = true

	b.writeStyleds(es.promptContent)

	if b.line() == 0 && b.col*2 < b.width {
		b.indent = b.col
	}

	// i keeps track of number of bytes written.
	i := 0

	// nowAt is called at every rune boundary.
	nowAt := func(i int) {
		if mode == modeCompletion && i == es.completion.begin {
			c := es.completion.selectedCandidate()
			b.writes(c.text, styleForCompleted.String())
		}
		if i == es.dot {
			b.dot = b.cursor()
		}
	}
	nowAt(0)
tokens:
	for _, token := range es.tokens {
		for _, r := range token.Text {
			if mode == modeCompletion &&
				es.completion.begin <= i && i <= es.completion.end {
				// Do nothing. This part is replaced by the completion candidate.
			} else {
				b.write(r, joinStyles(styleForType[token.Type], token.MoreStyle).String())
			}
			i += utf8.RuneLen(r)

			nowAt(i)
			if mode == modeHistory && i == len(es.hist.prefix) {
				break tokens
			}
		}
	}

	if mode == modeHistory {
		// Put the rest of current history, position the cursor at the
		// end of the line, and finish writing
		h := es.hist
		b.writes(h.line[len(h.prefix):], styleForCompletedHistory.String())
		b.dot = b.cursor()
	}

	// Write rprompt
	if len(es.rpromptContent) > 0 {
		padding := b.width - b.col
		for _, s := range es.rpromptContent {
			padding -= util.Wcswidth(s.text)
		}
		if padding >= 1 {
			b.newlineWhenFull = false
			b.writePadding(padding, "")
			b.writeStyleds(es.rpromptContent)
		}
	}

	// bufMode
	bufMode = render(es.mode.ModeLine(), width)

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(es.tips) > 0 {
		bufTips = newBuffer(width)
		bufTips.writes(strings.Join(es.tips, "\n"), styleForTip.String())
	}

	hListing := 0
	// Trim lines and determine the maximum height for bufListing
	// TODO come up with a UI to tell the user that something is not shown.
	switch {
	case height >= lines(bufNoti, bufLine, bufMode, bufTips):
		hListing = height - lines(bufLine, bufMode, bufTips)
	case height >= lines(bufNoti, bufLine, bufTips):
		bufMode = nil
	case height >= lines(bufNoti, bufLine):
		bufMode = nil
		if bufTips != nil {
			bufTips.trimToLines(0, height-lines(bufNoti, bufLine))
		}
	case height >= lines(bufLine):
		bufTips, bufMode = nil, nil
		if bufNoti != nil {
			n := len(bufNoti.cells)
			bufNoti.trimToLines(n-(height-lines(bufLine)), n)
		}
	case height >= 1:
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine+1-height, dotLine+1)
	default:
		// Broken terminal. Still try to render one line of bufLine.
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine, dotLine+1)
	}

	// bufListing.
	if hListing > 0 {
		if lister, ok := es.mode.(ListRenderer); ok {
			bufListing = lister.ListRender(width, hListing)
		} else if lister, ok := es.mode.(Lister); ok {
			bufListing = render(lister.List(hListing), width)
		}
		// XXX When in completion mode, we re-render the mode line, since the
		// scrollbar in the mode line depends on completion.lastShown which is
		// only known after the listing has been rendered. Since rendering the
		// scrollbar never adds additional lines to bufMode, we may do this
		// without recalculating the layout.
		if mode == modeCompletion {
			bufMode = render(es.mode.ModeLine(), width)
		}
	}

	if logWriterDetail {
		Logger.Printf("bufLine %d, bufMode %d, bufTips %d, bufListing %d",
			lines(bufLine), lines(bufMode), lines(bufTips), lines(bufListing))
	}

	// XXX
	buf.cells = nil
	// Combine buffers (reusing bufLine)
	buf.extend(bufLine, true)
	cursorOnModeLine := false
	if coml, ok := es.mode.(CursorOnModeLiner); ok {
		cursorOnModeLine = coml.CursorOnModeLine()
	}
	buf.extend(bufMode, cursorOnModeLine)
	buf.extend(bufTips, false)
	buf.extend(bufListing, false)

	er.bufNoti = bufNoti
}

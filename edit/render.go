package edit

import (
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
	b.writePadding(1, "")
	b.writes(ml.filter, styleForFilter.String())
	b.dot = b.cursor()
}

type modeLineWithScrollBarRenderer struct {
	modeLineRenderer
	n, low, high int
}

func (ml modeLineWithScrollBarRenderer) render(b *buffer) {
	ml.modeLineRenderer.render(b)

	scrollbarWidth := b.width - cellsWidth(b.lines[len(b.lines)-1]) - 2
	if scrollbarWidth >= 3 {
		b.writePadding(1, "")
		writeHorizontalScrollbar(b, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}

type placeholderRenderer string

func (lp placeholderRenderer) render(b *buffer) {
	b.writes(util.TrimWcwidth(string(lp), b.width), "")
}

type listingRenderer struct {
	lines []styled
}

func (ls listingRenderer) render(b *buffer) {
	for i, line := range ls.lines {
		if i > 0 {
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
	b.extendRight(b1, 0)

	scrollbar := renderScrollbar(ls.n, ls.low, ls.high, ls.height)
	b.extendRight(scrollbar, b.width-1)
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
	b.extendRight(bParent, 0)

	bCurrent := render(nr.current, wCurrent)
	b.extendRight(bCurrent, wParent+margin)

	if wPreview > 0 {
		bPreview := render(nr.preview, wPreview)
		b.extendRight(bPreview, wParent+wCurrent+2*margin)
	}
}

// linesRenderer renders lines with a uniform style.
type linesRenderer struct {
	lines []string
	style string
}

func (nr linesRenderer) render(b *buffer) {
	b.writes(strings.Join(nr.lines, "\n"), "")
}

// cmdlineRenderer renders the command line, including the prompt, the user's
// input and the rprompt.
type cmdlineRenderer struct {
	prompt  []*styled
	line    string
	styling *styling
	dot     int
	rprompt []*styled

	hasComp   bool
	compBegin int
	compEnd   int
	compText  string

	hasHist   bool
	histBegin int
	histText  string
}

func newCmdlineRenderer(p []*styled, l string, s *styling, d int, rp []*styled) *cmdlineRenderer {
	return &cmdlineRenderer{prompt: p, line: l, styling: s, dot: d, rprompt: rp}
}

func (clr *cmdlineRenderer) setComp(b, e int, t string) {
	clr.hasComp = true
	clr.compBegin, clr.compEnd, clr.compText = b, e, t
}

func (clr *cmdlineRenderer) setHist(b int, t string) {
	clr.hasHist = true
	clr.histBegin, clr.histText = b, t
}

func (clr *cmdlineRenderer) render(b *buffer) {
	b.eagerWrap = true

	b.writeStyleds(clr.prompt)

	// If the prompt takes less than half of a line, set the indent.
	if len(b.lines) == 1 && b.col*2 < b.width {
		b.indent = b.col
	}

	// i keeps track of number of bytes written.
	i := 0

	applier := clr.styling.apply()

	// nowAt is called at every rune boundary.
	nowAt := func(i int) {
		applier.at(i)
		if clr.hasComp && i == clr.compBegin {
			b.writes(clr.compText, styleForCompleted.String())
		}
		if i == clr.dot {
			b.dot = b.cursor()
		}
	}
	nowAt(0)

	for _, r := range clr.line {
		if clr.hasComp && clr.compBegin <= i && i < clr.compEnd {
			// Do nothing. This part is replaced by the completion candidate.
		} else {
			b.write(r, applier.get())
		}
		i += utf8.RuneLen(r)

		nowAt(i)
		if clr.hasHist && i == clr.histBegin {
			break
		}
	}

	if clr.hasHist {
		// Put the rest of current history and position the cursor at the
		// end of the line.
		b.writes(clr.histText, styleForCompletedHistory.String())
		b.dot = b.cursor()
	}

	// Write rprompt
	if len(clr.rprompt) > 0 {
		padding := b.width - b.col
		for _, s := range clr.rprompt {
			padding -= util.Wcswidth(s.text)
		}
		if padding >= 1 {
			b.eagerWrap = false
			b.writePadding(padding, "")
			b.writeStyleds(clr.rprompt)
		}
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
		bufNoti = render(linesRenderer{es.notifications, ""}, width)
		es.notifications = nil
	}

	// bufLine
	clr := newCmdlineRenderer(es.promptContent, es.line, es.styling, es.dot, es.rpromptContent)
	switch mode {
	case modeCompletion:
		c := es.completion
		clr.setComp(c.begin, c.end, c.selectedCandidate().code)
	case modeHistory:
		begin := len(es.hist.prefix)
		clr.setHist(begin, es.hist.line[begin:])
	}
	bufLine = render(clr, width)

	// bufMode
	bufMode = render(es.mode.ModeLine(), width)

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(es.tips) > 0 {
		bufTips = render(linesRenderer{es.tips, styleForTip.String()}, width)
	}

	hListing := 0
	// Trim lines and determine the maximum height for bufListing
	// TODO come up with a UI to tell the user that something is not shown.
	switch {
	case height >= buffersHeight(bufNoti, bufLine, bufMode, bufTips):
		hListing = height - buffersHeight(bufLine, bufMode, bufTips)
	case height >= buffersHeight(bufNoti, bufLine, bufTips):
		bufMode = nil
	case height >= buffersHeight(bufNoti, bufLine):
		bufMode = nil
		if bufTips != nil {
			bufTips.trimToLines(0, height-buffersHeight(bufNoti, bufLine))
		}
	case height >= buffersHeight(bufLine):
		bufTips, bufMode = nil, nil
		if bufNoti != nil {
			n := len(bufNoti.lines)
			bufNoti.trimToLines(n-(height-buffersHeight(bufLine)), n)
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
		logger.Printf("bufLine %d, bufMode %d, bufTips %d, bufListing %d",
			buffersHeight(bufLine), buffersHeight(bufMode), buffersHeight(bufTips), buffersHeight(bufListing))
	}

	// XXX
	buf.lines = nil
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

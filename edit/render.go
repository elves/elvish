package edit

import (
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

type modeLineRenderer struct {
	title  string
	filter string
}

func (ml modeLineRenderer) Render(b *ui.Buffer) {
	b.WriteString(ml.title, styleForMode.String())
	b.WriteSpaces(1, "")
	b.WriteString(ml.filter, styleForFilter.String())
	b.Dot = b.Cursor()
}

type modeLineWithScrollBarRenderer struct {
	modeLineRenderer
	n, low, high int
}

func (ml modeLineWithScrollBarRenderer) Render(b *ui.Buffer) {
	ml.modeLineRenderer.Render(b)

	scrollbarWidth := b.Width - ui.CellsWidth(b.Lines[len(b.Lines)-1]) - 2
	if scrollbarWidth >= 3 {
		b.WriteSpaces(1, "")
		writeHorizontalScrollbar(b, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}

type placeholderRenderer string

func (lp placeholderRenderer) Render(b *ui.Buffer) {
	b.WriteString(util.TrimWcwidth(string(lp), b.Width), "")
}

type listingRenderer struct {
	lines []ui.Styled
}

func (ls listingRenderer) Render(b *ui.Buffer) {
	for i, line := range ls.lines {
		if i > 0 {
			b.Newline()
		}
		b.WriteString(util.ForceWcwidth(line.Text, b.Width), line.Styles.String())
	}
}

type listingWithScrollBarRenderer struct {
	listingRenderer
	n, low, high, height int
}

func (ls listingWithScrollBarRenderer) Render(b *ui.Buffer) {
	b1 := ui.Render(ls.listingRenderer, b.Width-1)
	b.ExtendRight(b1, 0)

	scrollbar := renderScrollbar(ls.n, ls.low, ls.high, ls.height)
	b.ExtendRight(scrollbar, b.Width-1)
}

type navRenderer struct {
	maxHeight                      int
	fwParent, fwCurrent, fwPreview int
	parent, current, preview       ui.Renderer
}

func makeNavRenderer(h int, w1, w2, w3 int, r1, r2, r3 ui.Renderer) ui.Renderer {
	return &navRenderer{h, w1, w2, w3, r1, r2, r3}
}

const navColMargin = 1

func (nr *navRenderer) Render(b *ui.Buffer) {
	wParent, wCurrent, wPreview := getNavWidths(b.Width-navColMargin*2,
		nr.fwCurrent, nr.fwPreview)

	bParent := ui.Render(nr.parent, wParent)
	b.ExtendRight(bParent, 0)

	bCurrent := ui.Render(nr.current, wCurrent)
	b.ExtendRight(bCurrent, wParent+navColMargin)

	if wPreview > 0 {
		bPreview := ui.Render(nr.preview, wPreview)
		b.ExtendRight(bPreview, wParent+wCurrent+2*navColMargin)
	}
}

// linesRenderer renders lines with a uniform style.
type linesRenderer struct {
	lines []string
	style string
}

func (nr linesRenderer) Render(b *ui.Buffer) {
	b.WriteString(strings.Join(nr.lines, "\n"), "")
}

// cmdlineRenderer renders the command line, including the prompt, the user's
// input and the rprompt.
type cmdlineRenderer struct {
	prompt  []*ui.Styled
	line    string
	styling *highlight.Styling
	dot     int
	rprompt []*ui.Styled

	hasComp   bool
	compBegin int
	compEnd   int
	compText  string

	hasHist   bool
	histBegin int
	histText  string
}

func newCmdlineRenderer(p []*ui.Styled, l string, s *highlight.Styling, d int, rp []*ui.Styled) *cmdlineRenderer {
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

func (clr *cmdlineRenderer) Render(b *ui.Buffer) {
	b.EagerWrap = true

	b.WriteStyleds(clr.prompt)

	// If the prompt takes less than half of a line, set the indent.
	if len(b.Lines) == 1 && b.Col*2 < b.Width {
		b.Indent = b.Col
	}

	// i keeps track of number of bytes written.
	i := 0

	applier := clr.styling.Apply()

	// nowAt is called at every rune boundary.
	nowAt := func(i int) {
		applier.At(i)
		if clr.hasComp && i == clr.compBegin {
			b.WriteString(clr.compText, styleForCompleted.String())
		}
		if i == clr.dot {
			b.Dot = b.Cursor()
		}
	}
	nowAt(0)

	for _, r := range clr.line {
		if clr.hasComp && clr.compBegin <= i && i < clr.compEnd {
			// Do nothing. This part is replaced by the completion candidate.
		} else {
			b.Write(r, applier.Get())
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
		b.WriteString(clr.histText, styleForCompletedHistory.String())
		b.Dot = b.Cursor()
	}

	// Write rprompt
	if len(clr.rprompt) > 0 {
		padding := b.Width - b.Col
		for _, s := range clr.rprompt {
			padding -= util.Wcswidth(s.Text)
		}
		if padding >= 1 {
			b.EagerWrap = false
			b.WriteSpaces(padding, "")
			b.WriteStyleds(clr.rprompt)
		}
	}
}

var logEditorRender = false

// editorRenderer renders the entire editor.
type editorRenderer struct {
	*editorState
	height  int
	bufNoti *ui.Buffer
}

func (er *editorRenderer) Render(buf *ui.Buffer) {
	height, width, es := er.height, buf.Width, er.editorState

	var bufNoti, bufLine, bufMode, bufTips, bufListing *ui.Buffer
	// butNoti
	if len(es.notifications) > 0 {
		bufNoti = ui.Render(linesRenderer{es.notifications, ""}, width)
		es.notifications = nil
	}

	// bufLine
	clr := newCmdlineRenderer(es.promptContent, es.buffer, es.styling, es.dot, es.rpromptContent)
	// TODO(xiaq): Instead of doing a type switch, expose an API for modes to
	// modify the text (and mark their part as modified).
	switch mode := es.mode.(type) {
	case *completion:
		c := es.completion
		clr.setComp(c.begin, c.end, c.selectedCandidate().code)
	case *hist:
		begin := len(mode.Prefix())
		clr.setHist(begin, mode.CurrentCmd()[begin:])
	}
	bufLine = ui.Render(clr, width)

	// bufMode
	bufMode = ui.Render(es.mode.ModeLine(), width)

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(es.tips) > 0 {
		bufTips = ui.Render(linesRenderer{es.tips, styleForTip.String()}, width)
	}

	hListing := 0
	// Trim lines and determine the maximum height for bufListing
	// TODO come up with a UI to tell the user that something is not shown.
	switch {
	case height >= ui.BuffersHeight(bufNoti, bufLine, bufMode, bufTips):
		hListing = height - ui.BuffersHeight(bufLine, bufMode, bufTips)
	case height >= ui.BuffersHeight(bufNoti, bufLine, bufTips):
		bufMode = nil
	case height >= ui.BuffersHeight(bufNoti, bufLine):
		bufMode = nil
		if bufTips != nil {
			bufTips.TrimToLines(0, height-ui.BuffersHeight(bufNoti, bufLine))
		}
	case height >= ui.BuffersHeight(bufLine):
		bufTips, bufMode = nil, nil
		if bufNoti != nil {
			n := len(bufNoti.Lines)
			bufNoti.TrimToLines(n-(height-ui.BuffersHeight(bufLine)), n)
		}
	case height >= 1:
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.Dot.Line
		bufLine.TrimToLines(dotLine+1-height, dotLine+1)
	default:
		// Broken terminal. Still try to render one line of bufLine.
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.Dot.Line
		bufLine.TrimToLines(dotLine, dotLine+1)
	}

	// bufListing.
	if hListing > 0 {
		if lister, ok := es.mode.(ListRenderer); ok {
			bufListing = lister.ListRender(width, hListing)
		} else if lister, ok := es.mode.(Lister); ok {
			bufListing = ui.Render(lister.List(hListing), width)
		}
		// XXX When in completion mode, we re-render the mode line, since the
		// scrollbar in the mode line depends on completion.lastShown which is
		// only known after the listing has been rendered. Since rendering the
		// scrollbar never adds additional lines to bufMode, we may do this
		// without recalculating the layout.
		if _, ok := es.mode.(*completion); ok {
			bufMode = ui.Render(es.mode.ModeLine(), width)
		}
	}

	if logEditorRender {
		logger.Printf("bufLine %d, bufMode %d, bufTips %d, bufListing %d",
			ui.BuffersHeight(bufLine), ui.BuffersHeight(bufMode), ui.BuffersHeight(bufTips), ui.BuffersHeight(bufListing))
	}

	// XXX
	buf.Lines = nil
	// Combine buffers (reusing bufLine)
	buf.Extend(bufLine, true)
	cursorOnModeLine := false
	if coml, ok := es.mode.(CursorOnModeLiner); ok {
		cursorOnModeLine = coml.CursorOnModeLine()
	}
	buf.Extend(bufMode, cursorOnModeLine)
	buf.Extend(bufTips, false)
	buf.Extend(bufListing, false)

	er.bufNoti = bufNoti
}

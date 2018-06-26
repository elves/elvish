package edcore

import (
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

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

	hasRepl   bool
	replBegin int
	replEnd   int
	replText  string
}

func newCmdlineRenderer(p []*ui.Styled, l string, s *highlight.Styling, d int, rp []*ui.Styled) *cmdlineRenderer {
	return &cmdlineRenderer{prompt: p, line: l, styling: s, dot: d, rprompt: rp}
}

func (clr *cmdlineRenderer) setRepl(b, e int, t string) {
	clr.hasRepl = true
	clr.replBegin, clr.replEnd, clr.replText = b, e, t
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
		// Replacement should be written before setting b.Dot. This way, if the
		// replacement starts right at the dot, the cursor is correctly placed
		// after the replacement.
		if clr.hasRepl && i == clr.replBegin {
			b.WriteString(clr.replText, styleForReplacement.String())
		}
		if i == clr.dot {
			b.Dot = b.Cursor()
		}
	}
	nowAt(0)

	for _, r := range clr.line {
		if clr.hasRepl && clr.replBegin <= i && i < clr.replEnd {
			// Do nothing. This part is replaced by the replacement.
		} else {
			b.Write(r, applier.Get())
		}
		i += utf8.RuneLen(r)

		nowAt(i)
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
	if repl, ok := es.mode.(replacementer); ok {
		clr.setRepl(repl.Replacement())
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
		// Determine a window of bufLine that has $height lines around the line
		// where the dot is currently on.
		low := bufLine.Dot.Line - height/2
		high := low + height
		if low < 0 {
			low = 0
			high = low + height
		} else if high > len(bufLine.Lines) {
			high = len(bufLine.Lines)
			low = high - height
		}
		bufLine.TrimToLines(low, high)
	default:
		// Broken terminal. Still try to render one line of bufLine.
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.Dot.Line
		bufLine.TrimToLines(dotLine, dotLine+1)
	}

	// bufListing.
	if hListing > 0 {
		switch mode := es.mode.(type) {
		case listRenderer:
			bufListing = mode.ListRender(width, hListing)
		case lister:
			bufListing = ui.Render(mode.List(hListing), width)
		}
		// XXX When in completion mode, we re-render the mode line, since the
		// scrollbar in the mode line depends on completion.lastShown which is
		// only known after the listing has been rendered. Since rendering the
		// scrollbar never adds additional lines to bufMode, we may do this
		// without recalculating the layout.
		if _, ok := es.mode.(redrawModeLiner); ok {
			bufMode = ui.Render(es.mode.ModeLine(), width)
		}
	}

	if logEditorRender {
		logger.Printf("bufNoti %d, bufLine %d, bufMode %d, bufTips %d, "+
			"hListing %d, bufListing %d",
			ui.BuffersHeight(bufNoti), ui.BuffersHeight(bufLine),
			ui.BuffersHeight(bufMode), ui.BuffersHeight(bufTips),
			hListing, ui.BuffersHeight(bufListing))
	}

	// XXX
	buf.Lines = nil
	// Combine buffers (reusing bufLine)
	buf.Extend(bufLine, true)
	cursorOnModeLine := false
	if coml, ok := es.mode.(cursorOnModeLiner); ok {
		cursorOnModeLine = coml.CursorOnModeLine()
	}
	buf.Extend(bufMode, cursorOnModeLine)
	buf.Extend(bufTips, false)
	buf.Extend(bufListing, false)

	er.bufNoti = bufNoti
}

package etk

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

// View is the appearance of a component.
type View interface {
	// Render renders onto a region in the terminal bound by a rectangle.
	Render(width, height int) *term.Buffer
}

// EmptyView is a View that has no content and occupies no space.
type EmptyView struct{}

func (v EmptyView) Render(width, height int) *term.Buffer { return &term.Buffer{Width: width} }

type HorizontalGapView struct{ Width int }

func (v HorizontalGapView) Render(width, height int) *term.Buffer {
	width = min(width, v.Width)
	line := make([]term.Cell, width)
	for i := range line {
		line[i] = term.Cell{Text: " "}
	}
	return &term.Buffer{Width: width, Lines: [][]term.Cell{line}}
}

type VerticalGapView struct{ Height int }

func (v VerticalGapView) Render(width, height int) *term.Buffer {
	return &term.Buffer{Lines: make([][]term.Cell, min(height, v.Height))}
}

type TextWrap uint8

const (
	LazyWrap TextWrap = iota
	EagerWrap
	NoWrap
)

type TextView struct {
	Spans     []ui.Text
	DotBefore int
	Wrap      TextWrap
}

var DotHere ui.Text = ui.Text{}

func Text(spans ...ui.Text) TextView          { return makeText(LazyWrap, spans) }
func TextNoWrap(spans ...ui.Text) TextView    { return makeText(NoWrap, spans) }
func TextEagerWrap(spans ...ui.Text) TextView { return makeText(EagerWrap, spans) }

func makeText(w TextWrap, spans []ui.Text) TextView {
	v := TextView{make([]ui.Text, 0, len(spans)), 0, w}
	for i, span := range spans {
		if unsafe.SliceData(span) == unsafe.SliceData(DotHere) {
			v.DotBefore = i
		} else {
			v.Spans = append(v.Spans, span)
		}
	}
	return v
}

func (v TextView) Render(width, height int) *term.Buffer {
	buf := term.Buffer{Width: width}
	var line []term.Cell
	var col int
	newLine := func() {
		buf.Lines = append(buf.Lines, line)
		line = nil
		col = 0
	}
	for i, span := range v.Spans {
		for _, seg := range span {
			styleSGR := seg.Style.SGR()
			for _, r := range seg.Text {
				if r == '\n' {
					newLine()
					continue
				}
				cell := PrintCell(r, styleSGR)
				col += wcwidth.Of(cell.Text)
				if col > width {
					if v.Wrap == NoWrap {
						continue
					}
					newLine()
					line = append(line, cell)
				} else {
					line = append(line, cell)
					if col == width && v.Wrap == EagerWrap {
						newLine()
					}
				}
			}
		}
		if v.DotBefore == i+1 {
			buf.Dot = term.Pos{Line: len(buf.Lines), Col: col}
		}
	}
	buf.Lines = append(buf.Lines, line)
	if len(buf.Lines) > height {
		// TODO: Ensure that the dot is visible?
		buf.Lines = buf.Lines[:height]
	}
	return &buf
}

// PrintCell "prints" ASCII control characters by replacing them with their
// caret notation and adding reverse video. It doesn't handle other unprintable
// characters.
func PrintCell(r rune, sgr string) term.Cell {
	if r < 0x20 || r == 0x7f {
		text := "^" + string(r^0x40)
		if sgr != "" {
			sgr += ";7"
		} else {
			sgr = "7"
		}
		return term.Cell{Text: text, Style: sgr}
	}
	return term.Cell{Text: string(r), Style: sgr}
}

type BoxView struct {
	Children   []BoxChild
	Focus      int
	Horizontal bool
}

type BoxChild struct {
	View View
	Flex bool
}

// Child verbs
//  - "=" for non-flex child
//  - "*" for flex child
//
// Example:
//
// a* b= c*

// Expressing focus
//
// a* [b=] c*

// Expressing gap

// a* 2 [b=] 1 c*

// Calling box
//
// Box("a* [b=] c*", aView, bView, cView)
//
// Box(`a*
//      [b=]
//      c*`, aView, bView, cView)

func Box(layout string, views ...View) BoxView {
	return BoxFocus(layout, 0, views...)
}

func BoxFocus(layout string, focus int, views ...View) BoxView {
	// Parse the layout string.
	var verbs []string
	var horizontal bool
	lines := strings.Split(strings.TrimSpace(layout), "\n")
	if len(lines) == 1 {
		// Only one line: horizontal layout, each field is a verb
		verbs = strings.Fields(lines[0])
		horizontal = true
	} else {
		// Vertical layout, each line must contain one field as the verb
		verbs = make([]string, len(lines))
		for i, line := range lines {
			fields := strings.Fields(line)
			if len(fields) != 1 {
				panic(fmt.Sprintf("multi-line layout with %d fields: %q", len(fields), layout))
			}
			verbs[i] = fields[0]
		}
	}

	var children []BoxChild
	j := 0 // index into views
	for _, verb := range verbs {
		if strings.HasPrefix(verb, "[") {
			verb = verb[1 : len(verb)-1]
			focus = len(children)
		}
		var flex bool
		if strings.HasSuffix(verb, "*") {
			flex = true
		} else if strings.HasSuffix(verb, "=") {
		} else if n, err := strconv.Atoi(verb); err == nil {
			var view View
			if horizontal {
				view = HorizontalGapView{n}
			} else {
				view = VerticalGapView{n}
			}
			children = append(children, BoxChild{view, false})
			continue
		} else {
			panic(fmt.Sprintf("invalid verb (must be number or end in * or =): %q", verb))
		}
		children = append(children, BoxChild{views[j], flex})
		j++
	}
	if j != len(views) {
		panic(fmt.Sprintf("superfluous view: consumed %d out of %d", j, len(views)))
	}
	return BoxView{children, focus, horizontal}
}

func (v BoxView) Render(width, height int) *term.Buffer {
	var render func(View, bool) *term.Buffer
	var flexChildren int
	if v.Horizontal {
		budget := width
		render = func(v View, flex bool) *term.Buffer {
			width := budget
			if flex {
				width = budget / flexChildren
			}
			buf := v.Render(width, height)
			// TODO: Maybe term.Buffer should keep track of the actual width
			// itself
			if !flex {
				actualWidth := 0
				for _, line := range buf.Lines {
					actualWidth = max(actualWidth, cellsWidth(line))
				}
				buf.Width = actualWidth
			}
			budget -= buf.Width
			return buf
		}
	} else {
		budget := height
		render = func(v View, flex bool) *term.Buffer {
			height := budget
			if flex {
				height = budget / flexChildren
			}
			buf := v.Render(width, height)
			budget -= len(buf.Lines)
			return buf
		}
	}

	childBufs := make([]*term.Buffer, len(v.Children))
	// Render non-flex children first.
	for i, child := range v.Children {
		if child.Flex {
			flexChildren++
		} else {
			childBufs[i] = render(child.View, false)
		}
	}
	// Render flex children first.
	for i, child := range v.Children {
		if child.Flex {
			childBufs[i] = render(child.View, true)
			flexChildren--
		}
	}

	var buf term.Buffer
	for i, childBuf := range childBufs {
		moveDot := i == v.Focus
		if v.Horizontal {
			buf.ExtendRight(childBuf, moveDot)
		} else {
			buf.ExtendDown(childBuf, moveDot)
		}
	}
	return &buf
}

type ScrollBarView struct {
	Horizontal bool
	Total      int
	Low        int
	High       int
}

var (
	hscrollbarThumb  = ui.T(" ", ui.FgMagenta, ui.Inverse)
	hscrollbarTrough = ui.T("━", ui.FgMagenta)
	vscrollbarThumb  = ui.T(" ", ui.FgMagenta, ui.Inverse)
	vscrollbarTrough = ui.T("│", ui.FgMagenta)
)

func (v ScrollBarView) Render(width, height int) *term.Buffer {
	var budget int
	var thumbCell, troughCell term.Cell
	var buf term.Buffer
	var setCell func(i int, cell term.Cell)
	if v.Horizontal {
		budget = width
		thumbCell = segToCell(hscrollbarThumb[0])
		troughCell = segToCell(hscrollbarTrough[0])
		buf = term.Buffer{Width: width, Lines: [][]term.Cell{make([]term.Cell, width)}}
		setCell = func(i int, cell term.Cell) {
			buf.Lines[0][i] = cell
		}
	} else {
		budget = height
		thumbCell = segToCell(vscrollbarThumb[0])
		troughCell = segToCell(vscrollbarTrough[0])
		buf = term.Buffer{Width: 1, Lines: make([][]term.Cell, height)}
		setCell = func(i int, cell term.Cell) {
			buf.Lines[i] = []term.Cell{cell}
		}
	}

	posLow, posHigh := findScrollInterval(v.Total, v.Low, v.High, budget)
	for i := 0; i < budget; i++ {
		if posLow <= i && i < posHigh {
			setCell(i, thumbCell)
		} else {
			setCell(i, troughCell)
		}
	}
	return &buf
}

func segToCell(seg *ui.Segment) term.Cell {
	return term.Cell{Text: seg.Text, Style: seg.Style.SGR()}
}

func findScrollInterval(n, low, high, height int) (int, int) {
	f := func(i int) int {
		return int(float64(i)/float64(n)*float64(height) + 0.5)
	}
	scrollLow := f(low)
	// We use the following instead of f(high), so that the size of the
	// scrollbar remains the same as long as the window size remains the same.
	scrollHigh := scrollLow + f(high-low)

	if scrollLow == scrollHigh {
		if scrollHigh == height {
			scrollLow--
		} else {
			scrollHigh++
		}
	}
	return scrollLow, scrollHigh
}

func cellsWidth(cs []term.Cell) int {
	w := 0
	for _, c := range cs {
		w += wcwidth.Of(c.Text)
	}
	return w
}

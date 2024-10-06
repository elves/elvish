package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

type textAreaView struct {
	prompt  ui.Text
	rprompt ui.Text
	code    ui.Text
	dot     int
	tips    []ui.Text
}

var stylingForPending = ui.Underlined

func patchPending(c TextBuffer, p PendingText) (TextBuffer, int, int) {
	if p.From > p.To || p.From < 0 || p.To > len(c.Content) {
		// Invalid Pending.
		return c, 0, 0
	}
	if p.From == p.To && p.Content == "" {
		return c, 0, 0
	}
	newContent := c.Content[:p.From] + p.Content + c.Content[p.To:]
	newDot := 0
	switch {
	case c.Dot < p.From:
		// Dot is before the replaced region. Keep it.
		newDot = c.Dot
	case c.Dot >= p.From && c.Dot < p.To:
		// Dot is within the replaced region. Place the dot at the end.
		newDot = p.From + len(p.Content)
	case c.Dot >= p.To:
		// Dot is after the replaced region. Maintain the relative position of
		// the dot.
		newDot = c.Dot - (p.To - p.From) + len(p.Content)
	}
	return TextBuffer{Content: newContent, Dot: newDot}, p.From, p.From + len(p.Content)
}

func (v *textAreaView) Render(width, height int) *term.Buffer {
	bb := term.NewBufferBuilder(width)
	bb.EagerWrap = true

	bb.WriteStyled(v.prompt)
	if len(bb.Lines) == 1 && bb.Col*2 < bb.Width {
		bb.Indent = bb.Col
	}

	parts := v.code.Partition(v.dot)
	bb.
		WriteStyled(parts[0]).
		SetDotHere().
		WriteStyled(parts[1])

	bb.EagerWrap = false
	bb.Indent = 0

	// Handle rprompts with newlines.
	if rpromptWidth := styledWcswidth(v.rprompt); rpromptWidth > 0 {
		padding := bb.Width - bb.Col - rpromptWidth
		if padding >= 1 {
			bb.WriteSpaces(padding)
			bb.WriteStyled(v.rprompt)
		}
	}

	for _, tip := range v.tips {
		bb.Newline()
		bb.WriteStyled(tip)
	}

	b := bb.Buffer()
	truncateToHeight(b, height)
	return b
}

func truncateToHeight(b *term.Buffer, maxHeight int) {
	switch {
	case len(b.Lines) <= maxHeight:
		// We can show all line; do nothing.
	case b.Dot.Line < maxHeight:
		// We can show all lines before the cursor, and as many lines after the
		// cursor as we can, adding up to maxHeight.
		b.TrimToLines(0, maxHeight)
	default:
		// We can show maxHeight lines before and including the cursor line.
		b.TrimToLines(b.Dot.Line-maxHeight+1, b.Dot.Line+1)
	}
}

func styledWcswidth(t ui.Text) int {
	w := 0
	for _, seg := range t {
		w += wcwidth.Of(seg.Text)
	}
	return w
}

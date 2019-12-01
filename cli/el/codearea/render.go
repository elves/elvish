package codearea

import (
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
	"github.com/elves/elvish/util"
)

// View model, calculated from State and used for rendering.
type view struct {
	prompt  ui.Text
	rprompt ui.Text
	code    ui.Text
	dot     int
	errors  []error
}

var pendingStyle = ui.Underlined

func getView(w *widget) *view {
	s := w.CopyState()
	code, pFrom, pTo := patchPending(s.Buffer, s.Pending)
	styledCode, errors := w.Highlighter(code.Content)
	if pFrom < pTo {
		// Apply pendingStyle to [pFrom, pTo)
		parts := styledCode.Partition(pFrom, pTo)
		pending := ui.Transform(parts[1], pendingStyle)
		styledCode = parts[0].ConcatText(pending).ConcatText(parts[2])
	}

	var rprompt ui.Text
	if !s.HideRPrompt {
		rprompt = w.RPrompt()
	}

	return &view{w.Prompt(), rprompt, styledCode, code.Dot, errors}
}

func patchPending(c Buffer, p Pending) (Buffer, int, int) {
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
	return Buffer{Content: newContent, Dot: newDot}, p.From, p.From + len(p.Content)
}

func renderView(v *view, buf *term.BufferBuilder) {
	buf.EagerWrap = true

	buf.WriteStyled(v.prompt)
	if len(buf.Lines) == 1 && buf.Col*2 < buf.Width {
		buf.Indent = buf.Col
	}

	parts := v.code.Partition(v.dot)
	buf.
		WriteStyled(parts[0]).
		SetDotHere().
		WriteStyled(parts[1])

	buf.EagerWrap = false
	buf.Indent = 0

	// Handle rprompts with newlines.
	if rpromptWidth := styledWcswidth(v.rprompt); rpromptWidth > 0 {
		padding := buf.Width - buf.Col - rpromptWidth
		if padding >= 1 {
			buf.WriteSpaces(padding)
			buf.WriteStyled(v.rprompt)
		}
	}

	if len(v.errors) > 0 {
		for _, err := range v.errors {
			buf.Newline()
			buf.Write(err.Error())
		}
	}
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
		w += util.Wcswidth(seg.Text)
	}
	return w
}

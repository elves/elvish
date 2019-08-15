package codearea

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

// View model, calculated from State and used for rendering.
type view struct {
	prompt  styled.Text
	rprompt styled.Text
	code    styled.Text
	dot     int
	errors  []error
}

const pendingStyle = "underlined"

func getView(s *State, hl func(string) (styled.Text, []error)) *view {
	code, pFrom, pTo := patchPendingCode(s.CodeBuffer, s.PendingCode)
	styledCode, errors := hl(code.Content)
	if pFrom < pTo {
		// Apply pendingStyle to [pFrom, pTo)
		parts := styledCode.Partition(pFrom, pTo)
		pending := styled.Transform(parts[1], pendingStyle)
		styledCode = parts[0].ConcatText(pending).ConcatText(parts[2])
	}

	return &view{s.Prompt, s.RPrompt, styledCode, code.Dot, errors}
}

func patchPendingCode(c CodeBuffer, p PendingCode) (CodeBuffer, int, int) {
	if p.From > p.To {
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
	return CodeBuffer{Content: newContent, Dot: newDot}, p.From, p.From + len(p.Content)
}

func renderView(v *view, buf *ui.BufferBuilder) {
	buf.EagerWrap = true

	buf.WriteLegacyStyleds(v.prompt.ToLegacyType())
	if len(buf.Lines) == 1 && buf.Col*2 < buf.Width {
		buf.Indent = buf.Col
	}

	parts := v.code.Partition(v.dot)
	buf.WriteLegacyStyleds(parts[0].ToLegacyType())
	buf.Dot = buf.Cursor()
	buf.WriteLegacyStyleds(parts[1].ToLegacyType())

	buf.EagerWrap = false

	// Handle rprompts with newlines.
	if rpromptWidth := styledWcswidth(v.rprompt); rpromptWidth > 0 {
		padding := buf.Width - buf.Col - rpromptWidth
		if padding >= 1 {
			buf.WriteSpacesSGR(padding, "")
			buf.WriteLegacyStyleds(v.rprompt.ToLegacyType())
		}
	}

	if len(v.errors) > 0 {
		for _, err := range v.errors {
			buf.Newline()
			buf.WritePlain(err.Error())
		}
	}
}

func truncateToHeight(b *ui.Buffer, maxHeight int) {
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

func styledWcswidth(t styled.Text) int {
	w := 0
	for _, seg := range t {
		w += util.Wcswidth(seg.Text)
	}
	return w
}

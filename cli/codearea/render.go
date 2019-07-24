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

func render(v *view, buf *ui.BufferBuilder) {
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
			buf.WriteSpaces(padding, "")
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

func styledWcswidth(t styled.Text) int {
	w := 0
	for _, seg := range t {
		w += util.Wcswidth(seg.Text)
	}
	return w
}

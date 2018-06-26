package styled

import "github.com/elves/elvish/edit/ui"

func (t Text) ToLegacyType() []*ui.Styled {
	out := make([]*ui.Styled, len(t))
	for i, seg := range t {
		out[i] = seg.ToLegacyType()
	}
	return out
}

func (s Segment) ToLegacyType() *ui.Styled {
	legacy := &ui.Styled{Text: s.Text}
	addLegacyStyle := func(style string) {
		legacy.Styles = append(legacy.Styles, style)
	}
	if s.Style.Foreground != "" {
		addLegacyStyle(s.Style.Foreground)
	}
	if s.Style.Background != "" {
		addLegacyStyle("bg-" + s.Style.Background)
	}
	if s.Style.Bold {
		addLegacyStyle("bold")
	}
	if s.Style.Dim {
		addLegacyStyle("dim")
	}
	if s.Style.Italic {
		addLegacyStyle("italic")
	}
	if s.Style.Underlined {
		addLegacyStyle("underlined")
	}
	if s.Style.Blink {
		addLegacyStyle("blink")
	}
	if s.Style.Inverse {
		addLegacyStyle("inverse")
	}
	return legacy
}

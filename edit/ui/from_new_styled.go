package ui

import "github.com/elves/elvish/styled"

// FromNewStyledText converts a styled.Text to a slice of *Styled.
func FromNewStyledText(t styled.Text) []*Styled {
	out := make([]*Styled, len(t))
	for i, seg := range t {
		out[i] = FromNewStyledSegment(seg)
	}
	return out
}

// FromNewStyledSegment converts a *styled.Segment to a *Styled.
func FromNewStyledSegment(s *styled.Segment) *Styled {
	legacy := &Styled{Text: s.Text}
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

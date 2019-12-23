package ui

import (
	"bytes"
	"fmt"
	"strings"
)

// String returns a string representation of the styled text. This now always
// assumes VT-style terminal output.
//
// TODO: Make string conversion sensible to environment, e.g. use HTML when
// output is web.
func (t Text) String() string {
	return t.VTString()
}

// VTString renders the styled text using VT-style escape sequences.
func (t Text) VTString() string {
	var buf bytes.Buffer
	for _, seg := range t {
		buf.WriteString(seg.VTString())
	}
	return buf.String()
}

// String returns a string representation of the styled segment. This now always
// assumes VT-style terminal output.
// TODO: Make string conversion sensible to environment, e.g. use HTML when
// output is web.
func (s *Segment) String() string {
	return s.VTString()
}

// VTString renders the styled segment using VT-style escape sequences.
func (s *Segment) VTString() string {
	sgr := s.SGR()
	if sgr == "" {
		return s.Text
	}
	return fmt.Sprintf("\033[%sm%s\033[m", sgr, s.Text)
}

// SGR returns SGR sequence for the style.
func (s Style) SGR() string {
	var sgr []string

	addIf := func(b bool, code string) {
		if b {
			sgr = append(sgr, code)
		}
	}
	addIf(s.Bold, "1")
	addIf(s.Dim, "2")
	addIf(s.Italic, "3")
	addIf(s.Underlined, "4")
	addIf(s.Blink, "5")
	addIf(s.Inverse, "7")
	if s.Foreground != nil {
		sgr = append(sgr, s.Foreground.fgSGR())
	}
	if s.Background != nil {
		sgr = append(sgr, s.Background.bgSGR())
	}

	return strings.Join(sgr, ";")
}

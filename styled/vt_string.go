package styled

import (
	"bytes"
	"fmt"
	"strings"
)

// String returns a string representation of the styled text. This now always
// assumes VT-style terminal output.
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
	sgr := make([]string, 0, 8)

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
	if code, ok := fgSGR[s.Foreground]; ok {
		sgr = append(sgr, code)
	}
	if code, ok := bgSGR[s.Background]; ok {
		sgr = append(sgr, code)
	}

	if len(sgr) == 0 {
		return s.Text
	}
	return fmt.Sprintf("\033[%sm%s\033[m", strings.Join(sgr, ";"), s.Text)
}

var fgSGR = map[string]string{
	"black":        "30",
	"red":          "31",
	"green":        "32",
	"yellow":       "33",
	"blue":         "34",
	"magenta":      "35",
	"cyan":         "36",
	"lightgray":    "37",
	"gray":         "90",
	"lightred":     "91",
	"lightgreen":   "92",
	"lightyellow":  "93",
	"lightblue":    "94",
	"lightmagenta": "95",
	"lightcyan":    "96",
	"white":        "97",
}

var bgSGR = map[string]string{
	"black":        "40",
	"red":          "41",
	"green":        "42",
	"yellow":       "43",
	"blue":         "44",
	"magenta":      "45",
	"cyan":         "46",
	"lightgray":    "47",
	"gray":         "100",
	"lightred":     "101",
	"lightgreen":   "102",
	"lightyellow":  "103",
	"lightblue":    "104",
	"lightmagenta": "105",
	"lightcyan":    "106",
	"white":        "107",
}

package ui

import (
	"fmt"
	"strings"
)

// NoColor can be set to true to suppress foreground and background colors when
// writing text to the terminal.
var NoColor bool = false

// Style specifies how something (mostly a string) shall be displayed.
type Style struct {
	Fg         Color
	Bg         Color
	Bold       bool
	Dim        bool
	Italic     bool
	Underlined bool
	Blink      bool
	Inverse    bool
}

// SGRValues returns an array of the individual SGR values for the style.
func (s Style) SGRValues() []string {
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
	if s.Fg != nil && !NoColor {
		sgr = append(sgr, s.Fg.fgSGR())
	}
	if s.Bg != nil && !NoColor {
		sgr = append(sgr, s.Bg.bgSGR())
	}
	return sgr
}

// SGR returns, for the Style, a string that can be included in an ANSI X3.64 SGR sequence.
func (s Style) SGR() string {
	return strings.Join(s.SGRValues(), ";")
}

// MergeFromOptions merges all recognized values from a map to the current
// Style.
func (s *Style) MergeFromOptions(options map[string]any) error {
	assignColor := func(val any, colorField *Color) string {
		if val == "default" {
			*colorField = nil
			return ""
		} else if s, ok := val.(string); ok {
			color := parseColor(s)
			if color != nil {
				*colorField = color
				return ""
			}
		}
		return "valid color string"
	}
	assignBool := func(val any, attrField *bool) string {
		if b, ok := val.(bool); ok {
			*attrField = b
		} else {
			return "bool value"
		}
		return ""
	}

	for k, v := range options {
		var need string

		switch k {
		case "fg-color":
			need = assignColor(v, &s.Fg)
		case "bg-color":
			need = assignColor(v, &s.Bg)
		case "bold":
			need = assignBool(v, &s.Bold)
		case "dim":
			need = assignBool(v, &s.Dim)
		case "italic":
			need = assignBool(v, &s.Italic)
		case "underlined":
			need = assignBool(v, &s.Underlined)
		case "blink":
			need = assignBool(v, &s.Blink)
		case "inverse":
			need = assignBool(v, &s.Inverse)

		default:
			return fmt.Errorf("unrecognized option '%s'", k)
		}

		if need != "" {
			return fmt.Errorf("value for option '%s' must be a %s", k, need)
		}
	}

	return nil
}

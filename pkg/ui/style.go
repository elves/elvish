package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// Style specifies how something (mostly a string) shall be displayed.
type Style struct {
	Foreground Color
	Background Color
	Bold       bool
	Dim        bool
	Italic     bool
	Underlined bool
	Blink      bool
	Inverse    bool
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

// MergeFromOptions merges all recognized values from a map to the current
// Style.
func (s *Style) MergeFromOptions(options map[string]interface{}) error {
	assignColor := func(val interface{}, colorField *Color) string {
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
	assignBool := func(val interface{}, attrField *bool) string {
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
			need = assignColor(v, &s.Foreground)
		case "bg-color":
			need = assignColor(v, &s.Background)
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

var sgrStyling = map[int]Styling{
	1: Bold,
	2: Dim,
	4: Underlined,
	5: Blink,
	7: Inverse,
}

// StyleFromStyling builds a Style from a Styling
func StyleFromStyling(s Styling) Style {
	var ret Style
	s.Transform(&ret)
	return ret
}

// StyleFromSGR builds a Style from an SGR sequence.
func StyleFromSGR(s string) Style {
	return StyleFromStyling(StylingFromSGR(s))
}

// StylingFromSGR builds a Styling from an SGR sequence.
func StylingFromSGR(s string) Styling {
	style := jointStyling{}
	codes := getSGRCodes(s)
	for len(codes) > 0 {
		code := codes[0]
		consume := 1

		switch {
		case sgrStyling[code] != nil:
			style = append(style, sgrStyling[code])
		case code == 0:
			style = append(style, resetStyling{})
		case 30 <= code && code <= 37:
			style = append(style, setForeground{ansiColor(code - 30)})
		case 40 <= code && code <= 47:
			style = append(style, setBackground{ansiColor(code - 40)})
		case 90 <= code && code <= 97:
			style = append(style, setForeground{ansiBrightColor(code - 90)})
		case 100 <= code && code <= 107:
			style = append(style, setBackground{ansiBrightColor(code - 100)})
		case code == 38 && len(codes) >= 3 && codes[1] == 5:
			style = append(style, setForeground{xterm256Color(codes[2])})
			consume = 3
		case code == 48 && len(codes) >= 3 && codes[1] == 5:
			style = append(style, setBackground{xterm256Color(codes[2])})
			consume = 3
		case code == 38 && len(codes) >= 5 && codes[1] == 2:
			style = append(style, setForeground{trueColor{uint8(codes[2]), uint8(codes[3]), uint8(codes[4])}})
			consume = 5
		case code == 48 && len(codes) >= 5 && codes[1] == 2:
			style = append(style, setBackground{trueColor{uint8(codes[2]), uint8(codes[3]), uint8(codes[4])}})
			consume = 5
		default:
			// Do nothing; skip this code
		}

		codes = codes[consume:]
	}
	return style
}

func getSGRCodes(s string) []int {
	var codes []int
	for _, part := range strings.Split(s, ";") {
		code, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		codes = append(codes, code)
	}
	return codes
}

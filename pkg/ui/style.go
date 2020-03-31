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

// ImportFromOptions assigns all recognized values from a map to the current
// Style.
func (s *Style) ImportFromOptions(options map[string]interface{}) error {
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
			return fmt.Errorf("value to option '%s' must be a %s", k, need)
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

// StyleFromSGR builds a Style from an SGR sequence.
func StyleFromSGR(s string) Style {
	style := Style{}
	codes := getSGRCodes(s)
	for len(codes) > 0 {
		code := codes[0]
		consume := 1

		switch {
		case sgrStyling[code] != nil:
			sgrStyling[code].transform(&style)
		case 30 <= code && code <= 37:
			style.Foreground = ansiColor(code - 30)
		case 40 <= code && code <= 47:
			style.Background = ansiColor(code - 40)
		case 90 <= code && code <= 97:
			style.Foreground = ansiBrightColor(code - 90)
		case 100 <= code && code <= 107:
			style.Background = ansiBrightColor(code - 100)
		case code == 38 && len(codes) >= 3 && codes[1] == 5:
			style.Foreground = xterm256Color(codes[2])
			consume = 3
		case code == 48 && len(codes) >= 3 && codes[1] == 5:
			style.Background = xterm256Color(codes[2])
			consume = 3
		case code == 38 && len(codes) >= 5 && codes[1] == 2:
			style.Foreground = trueColor{
				uint8(codes[2]), uint8(codes[3]), uint8(codes[4])}
			consume = 5
		case code == 48 && len(codes) >= 5 && codes[1] == 2:
			style.Background = trueColor{
				uint8(codes[2]), uint8(codes[3]), uint8(codes[4])}
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

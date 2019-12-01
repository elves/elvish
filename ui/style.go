package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// Style specifies how something (mostly a string) shall be displayed.
type Style struct {
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Italic     bool
	Underlined bool
	Blink      bool
	Inverse    bool
}

// ImportFromOptions assigns all recognized values from a map to the current
// Style.
func (s *Style) ImportFromOptions(options map[string]interface{}) error {
	assignColor := func(val interface{}, colorField *string) string {
		if val == "default" {
			*colorField = ""
		} else if c, ok := val.(string); ok && isValidColorName(c) {
			*colorField = c
		} else {
			return "valid color string"
		}
		return ""
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

func isValidColorName(col string) bool {
	switch col {
	case
		"black",
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"lightgray",
		"gray",
		"lightred",
		"lightgreen",
		"lightyellow",
		"lightblue",
		"lightmagenta",
		"lightcyan",
		"white":
		return true

	default:
		return false
	}
}

var sgrStyling = map[int]Styling{
	1:  Bold,
	2:  Dim,
	4:  Underlined,
	5:  Blink,
	7:  Inverse,
	30: Black,
	31: Red,
	32: Green,
	33: Yellow,
	34: Blue,
	35: Magenta,
	36: Cyan,
	37: White,
	40: BgBlack,
	41: BgRed,
	42: BgGreen,
	43: BgYellow,
	44: BgBlue,
	45: BgMagenta,
	46: BgCyan,
	47: BgLightGray,
}

// StyleFromSGR builds a Style from a ECMA-48 Set Graphics Rendition sequence, .
// a semicolon-delimited string of valid attribute codes Invalid codes are     .
// ignored                                                                     .
func StyleFromSGR(s string) Style {
	style := Style{}
	for _, part := range strings.Split(s, ";") {
		code, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		if styling, ok := sgrStyling[code]; ok {
			styling.transform(&style)
		}
	}
	return style
}

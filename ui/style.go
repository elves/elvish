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

var sgrStyling = map[int]string{
	1:  "bold",
	2:  "dim",
	4:  "underlined",
	5:  "blink",
	7:  "inverse",
	30: "black",
	31: "red",
	32: "green",
	33: "brown",
	34: "blue",
	35: "magenta",
	36: "cyan",
	37: "white",
	40: "bg-black",
	41: "bg-red",
	42: "bg-green",
	43: "bg-brown",
	44: "bg-blue",
	45: "bg-magenta",
	46: "bg-cyan",
	47: "bg-lightgray",
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
		if transform, ok := sgrStyling[code]; ok {
			FindStyling(transform)(&style)
		}
	}
	return style
}

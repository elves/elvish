package ui

import (
	"strings"

	"github.com/elves/elvish/styled"
)

var (
	sgrForForeground = map[string]string{
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
	sgrForBackground = map[string]string{
		"default":      "49",
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
)

const (
	sgrForBold       = "1"
	sgrForDim        = "2"
	sgrForItalic     = "3"
	sgrForUnderlined = "4"
	sgrForBlink      = "5"
	sgrForInverse    = "7"
)

func sgrFromStyle(s styled.Style) string {
	var codes []string
	add := func(code string) {
		if code != "" {
			codes = append(codes, code)
		}
	}
	if s.Foreground != "" {
		add(sgrForForeground[s.Foreground])
	}
	if s.Background != "" {
		add(sgrForBackground[s.Background])
	}
	if s.Bold {
		add(sgrForBold)
	}
	if s.Dim {
		add(sgrForDim)
	}
	if s.Italic {
		add(sgrForItalic)
	}
	if s.Underlined {
		add(sgrForUnderlined)
	}
	if s.Blink {
		add(sgrForBlink)
	}
	if s.Inverse {
		add(sgrForInverse)
	}
	return strings.Join(codes, ";")
}

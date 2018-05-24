package styled

import (
	"strings"
)

// todo: Make string conversion variable to environment.
// E.g. the shell displays colors different than HTML.
func (t Text) String() string {
	buf := make([]byte, 0, 64)
	for _, segment := range t {
		styleBuf := make([]string, 0, 8)

		if segment.bold != nil && *segment.bold {
			styleBuf = append(styleBuf, "1")
		}
		if segment.dim != nil && *segment.dim {
			styleBuf = append(styleBuf, "2")
		}
		if segment.italic != nil && *segment.italic {
			styleBuf = append(styleBuf, "3")
		}
		if segment.underlined != nil && *segment.underlined {
			styleBuf = append(styleBuf, "4")
		}
		if segment.blink != nil && *segment.blink {
			styleBuf = append(styleBuf, "5")
		}
		if segment.inverse != nil && *segment.inverse {
			styleBuf = append(styleBuf, "7")
		}

		if segment.Foreground != "default" {
			colString := segment.Foreground
			if col, ok := colorTranslationTable[colString]; ok {
				styleBuf = append(styleBuf, col)
			}
		}
		if segment.Background != "default" {
			colString := "bg-" + segment.Background
			if col, ok := colorTranslationTable[colString]; ok {
				styleBuf = append(styleBuf, col)
			}
		}

		if len(styleBuf) > 0 {
			buf = append(buf, "\033["...)
			buf = append(buf, strings.Join(styleBuf, ";")...)
			buf = append(buf, 'm')
			buf = append(buf, segment.Text...)
			buf = append(buf, "\033[m"...)
		} else {
			buf = append(buf, segment.Text...)
		}
	}
	return string(buf)
}

var colorTranslationTable = map[string]string{
	"default":      "39",
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

	"bg-default":      "49",
	"bg-black":        "40",
	"bg-red":          "41",
	"bg-green":        "42",
	"bg-yellow":       "43",
	"bg-blue":         "44",
	"bg-magenta":      "45",
	"bg-cyan":         "46",
	"bg-lightgray":    "47",
	"bg-gray":         "100",
	"bg-lightred":     "101",
	"bg-lightgreen":   "102",
	"bg-lightyellow":  "103",
	"bg-lightblue":    "104",
	"bg-lightmagenta": "105",
	"bg-lightcyan":    "106",
	"bg-white":        "107",
}

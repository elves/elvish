package ui

import (
	"strings"

	"github.com/elves/elvish/parse"
)

// Styled is a piece of text with style.
type Styled struct {
	Text   string
	Styles Styles
}

var styleTranslationTable = map[string]string{
	"bold":       "1",
	"dim":        "2",
	"italic":     "3",
	"underlined": "4",
	"blink":      "5",
	"inverse":    "7",

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

func Unstyled(s string) Styled {
	return Styled{s, Styles{}}
}

func (s *Styled) Kind() string {
	return "styled"
}

func (s *Styled) String() string {
	return "\033[" + s.Styles.String() + "m" + s.Text + "\033[m"
}

func (s *Styled) Repr(indent int) string {
	return "(le:styled " + parse.Quote(s.Text) + " " + parse.Quote(s.Styles.String()) + ")"
}

type Styles []string

func JoinStyles(so Styles, st ...Styles) Styles {
	for _, v := range st {
		so = append(so, v...)
	}

	return so
}

func TranslateStyle(s string) string {
	v, ok := styleTranslationTable[s]
	if ok {
		return v
	}
	return s
}

func StylesFromString(s string) Styles {
	var st Styles
	for _, v := range strings.Split(s, ";") {
		st = append(st, v)
	}

	return st
}

func (s Styles) String() string {
	var o string
	for i, v := range s {
		if len(v) > 0 {
			if i > 0 {
				o += ";"
			}
			o += TranslateStyle(v)
		}
	}

	return o
}

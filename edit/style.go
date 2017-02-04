package edit

import "strings"

// Styles for UI.
var (
	//styleForPrompt           = ""
	//styleForRPrompt          = "inverse"
	styleForCompleted        = styles{"dim"}
	styleForCompletedHistory = styles{"dim"}
	styleForMode             = styles{"bold", "lightgray", "bg-magenta"}
	styleForTip              = styles{}
	styleForFilter           = styles{"underlined"}
	styleForSelected         = styles{"inverse"}
	styleForScrollBarArea    = styles{"magenta"}
	styleForScrollBarThumb   = styles{"magenta", "inverse"}

	// Use default style for completion listing
	styleForCompletion = styles{}
	// Use inverse style for selected completion entry
	styleForSelectedCompletion = styles{"inverse"}
)

var styleForType = map[TokenKind]styles{
	ParserError:  styles{"bg-red"},
	Bareword:     styles{},
	SingleQuoted: styles{"yellow"},
	DoubleQuoted: styles{"yellow"},
	Variable:     styles{"magenta"},
	Wildcard:     styles{},
	Tilde:        styles{},
	Sep:          styles{},
}

var styleForSep = map[string]string{
	// unknown : "red",
	"#": "cyan",

	">":  "green",
	">>": "green",
	"<":  "green",
	"?>": "green",
	"|":  "green",

	"?(": "bold",
	"(":  "bold",
	")":  "bold",
	"[":  "bold",
	"]":  "bold",
	"{":  "bold",
	"}":  "bold",

	"&": "bold",

	"if":   "yellow",
	"then": "yellow",
	"elif": "yellow",
	"else": "yellow",
	"fi":   "yellow",

	"while": "yellow",
	"do":    "yellow",
	"done":  "yellow",

	"for": "yellow",
	"in":  "yellow",

	"try":     "yellow",
	"except":  "yellow",
	"finally": "yellow",
	"tried":   "yellow",

	"begin": "yellow",
	"end":   "yellow",
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

// Styles for semantic coloring.
var (
	styleForGoodCommand   = styles{"green"}
	styleForBadCommand    = styles{"red"}
	styleForBadVariable   = styles{"bg-red"}
	styleForCompilerError = styles{"bg-red"}
)

type styles []string

func joinStyles(so styles, st ...styles) styles {
	for _, v := range st {
		so = append(so, v...)
	}

	return so
}

func styleTranslated(s string) string {
	v, ok := styleTranslationTable[s]
	if ok {
		return v
	}
	return s
}

func stylesFromString(s string) styles {
	var st styles
	for _, v := range strings.Split(s, ";") {
		st = append(st, v)
	}

	return st
}

func (s styles) String() string {
	var o string
	for i, v := range s {
		if len(v) > 0 {
			if i > 0 {
				o += ";"
			}
			o += styleTranslated(v)
		}
	}

	return o
}

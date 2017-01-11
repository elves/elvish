package edit

import "strings"

// Styles for UI.
var (
	//styleForPrompt           = ""
	//styleForRPrompt          = "inverse"
	styleForCompleted        = styles{"dim"}
	styleForMode             = styles{"bold", "lightgray", "magenta"}
	styleForTip              = styles{}
	styleForCompletedHistory = styles{"dim"}
	styleForFilter           = styles{"underlined"}
	styleForSelected         = styles{"inverse"}
	styleForScrollBarArea    = styles{"magenta"}
	styleForScrollBarThumb   = styles{"magenta", "inverse"}
	styleForSideArrow        = styles{"inverse"}

	// Use black text on white for completion listing.
	styleForCompletion = styles{"black", "bg_white"}
	// Use white text on black for selected completion.
	styleForSelectedCompletion = "inverse"
)

var styleForType = map[TokenKind]styles{
	ParserError:  styles{"red", "3"},
	Bareword:     styles{},
	SingleQuoted: styles{"red"},
	DoubleQuoted: styles{"red"},
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
	//"unknown": "3",

	"bold":       "1",
	"dim":        "2",
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

	"bg_default":      "49",
	"bg_black":        "40",
	"bg_red":          "41",
	"bg_green":        "42",
	"bg_yellow":       "43",
	"bg_blue":         "44",
	"bg_magenta":      "45",
	"bg_cyan":         "46",
	"bg_lightgray":    "47",
	"bg_gray":         "100",
	"bg_lightred":     "101",
	"bg_lightgreen":   "102",
	"bg_lightyellow":  "103",
	"bg_lightblue":    "104",
	"bg_lightmagenta": "105",
	"bg_lightcyan":    "106",
	"bg_white":        "107",
}

// Styles for semantic coloring.
var (
	styleForGoodCommand   = styles{"green"}
	styleForBadCommand    = styles{"red"}
	styleForBadVariable   = styles{"red", "3"}
	styleForCompilerError = styles{"red", "3"}
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

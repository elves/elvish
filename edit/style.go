package edit

// Styles for UI.
var (
	styleForPrompt           = ""
	styleForRPrompt          = "7"
	styleForCompleted        = ";2"
	styleForMode             = "1;37;45"
	styleForTip              = ""
	styleForCompletedHistory = "2"
	styleForFilter           = "4"
	styleForSelected         = ";7"
)

var styleForType = map[TokenKind]string{
	ParserError:  "31;3",
	Bareword:     "",
	SingleQuoted: "33",
	DoubleQuoted: "33",
	Variable:     "35",
	Wildcard:     "",
	Tilde:        "",
	Sep:          "",
}

var styleForSep = map[string]string{
	// unknown : "31",
	"#": "36",

	">":  "32",
	"<":  "32",
	"?>": "32",
	"|":  "32",

	"?(": "1",
	"(":  "1",
	")":  "1",
	"[":  "1",
	"]":  "1",
	"{":  "1",
	"}":  "1",

	"&": "1",

	"if":    "33",
	"then":  "33",
	"elif":  "33",
	"else":  "33",
	"fi":    "33",
	"while": "33",
	"do":    "33",
	"done":  "33",
	"for":   "33",
	"in":    "33",
	"begin": "33",
	"end":   "33",
}

// Styles for semantic coloring.
var (
	styleForGoodCommand   = ";32"
	styleForBadCommand    = ";31"
	styleForBadVariable   = ";31;3"
	styleForCompilerError = ";31;3"
)

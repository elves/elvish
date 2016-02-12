package edit

// Styles for UI.
var (
	styleForPrompt            = ""
	styleForRPrompt           = "7"
	styleForCompleted         = ";4"
	styleForMode              = "1;3;35"
	styleForTip               = ""
	styleForCurrentCompletion = ";7"
	styleForCompletedHistory  = "4"
	styleForSelectedFile      = ";7"
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
}

// Styles for semantic coloring.
var (
	styleForGoodCommand   = ";32"
	styleForBadCommand    = ";31"
	styleForBadVariable   = ";31;3"
	styleForCompilerError = ";31;3"
)

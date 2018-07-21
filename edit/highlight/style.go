package highlight

import (
	"github.com/elves/elvish/parse"
)

// Semantically applied styles.
var (
	styleForGoodCommand  = "green"
	styleForBadCommand   = "red"
	styleForGoodVariable = "magenta"
)

// Lexically applied styles.

// Styles for Primary nodes.
var styleForPrimary = map[parse.PrimaryType]string{
	parse.Bareword:     "",
	parse.SingleQuoted: "yellow",
	parse.DoubleQuoted: "yellow",
	parse.Variable:     styleForGoodVariable,
	parse.Wildcard:     "",
	parse.Tilde:        "",
}

var styleForComment = "cyan"

// Styles for Sep nodes.
var styleForSep = map[string]string{
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

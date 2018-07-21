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

	"elif":    "yellow",
	"else":    "yellow",
	"except":  "yellow",
	"finally": "yellow",
}

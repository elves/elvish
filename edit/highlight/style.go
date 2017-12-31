package highlight

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
)

// Semantically applied styles.
var (
	styleForGoodCommand  = ui.Styles{"green"}
	styleForBadCommand   = ui.Styles{"red"}
	styleForGoodVariable = ui.Styles{"magenta"}
)

// Lexically applied styles.

// ui.Styles for Primary nodes.
var styleForPrimary = map[parse.PrimaryType]ui.Styles{
	parse.Bareword:     {},
	parse.SingleQuoted: {"yellow"},
	parse.DoubleQuoted: {"yellow"},
	parse.Variable:     styleForGoodVariable,
	parse.Wildcard:     {},
	parse.Tilde:        {},
}

var styleForComment = ui.Styles{"cyan"}

// ui.Styles for Sep nodes.
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

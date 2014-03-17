package edit

import (
	"github.com/xiaq/elvish/parse"
)

var (
	attrForPrompt            = ""
	attrForRprompt           = "7"
	attrForCompleted         = ";4"
	attrForMode              = "1;7;33"
	attrForTip               = ""
	attrForCurrentCompletion = ";7"
	attrForCompletedHistory  = "4"
	attrForSelectedFile      = ";7"
)

var attrForType = map[parse.ItemType]string{
	parse.ItemSpace:             "36", // only applies to comments
	parse.ItemSingleQuoted:      "33",
	parse.ItemDoubleQuoted:      "33",
	parse.ItemRedirLeader:       "32",
	parse.ItemStatusRedirLeader: "32",
	parse.ItemPipe:              "32",
	parse.ItemError:             "31",
	parse.ItemQuestionLParen:    "34;1",
	parse.ItemLParen:            "34;1",
	parse.ItemRParen:            "34;1",
	parse.ItemLBracket:          "34;1",
	parse.ItemRBracket:          "34;1",
	parse.ItemLBrace:            "34;1",
	parse.ItemRBrace:            "34;1",
	parse.ItemAmpersand:         "1",
	parse.ItemDollar:            "35",

	ItemValidCommand:    "32",
	ItemInvalidCommand:  "31",
	ItemValidVariable:   "35",
	ItemInvalidVariable: "31",
}

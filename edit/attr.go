package edit

import (
	"../parse"
)

var attrForType = map[parse.ItemType]string{
	parse.ItemSpace: "36", // only applies to comments
	parse.ItemSingleQuoted: "33",
	parse.ItemDoubleQuoted: "33",
	parse.ItemRedirLeader: "32",
	parse.ItemPipe: "32",
	parse.ItemError: "31",
	parse.ItemLParen: "34;1",
	parse.ItemRParen: "34;1",
	parse.ItemLBracket: "34;1",
	parse.ItemRBracket: "34;1",
	parse.ItemLBrace: "34;1",
	parse.ItemRBrace: "34;1",
	parse.ItemAmpersand: "1",
}

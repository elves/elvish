package edit

import (
	"../parse"
)

var attrForType = map[parse.ItemType]string{
	parse.ItemComment: "36",
	parse.ItemSingleQuoted: "33",
	parse.ItemDoubleQuoted: "33",
	parse.ItemRedirLeader: "32",
	parse.ItemPipe: "32",
	parse.ItemError: "31",
	parse.ItemLParen: "34;1",
	parse.ItemRParen: "34;1",
}

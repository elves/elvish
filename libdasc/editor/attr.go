package editor

import (
	"../parse"
)

var attrForType = map[parse.ItemType]string{
	parse.ItemSingleQuoted: "33",
	parse.ItemDoubleQuoted: "33",
	parse.ItemError: "31",
}

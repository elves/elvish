package main

import (
	"strings"

	"src.elv.sh/pkg/md"
)

func addImplicitElvdocTargets(module string, ops []md.InlineOp) {
	for i := range ops {
		if i+2 < len(ops) &&
			ops[i].Type == md.OpLinkStart && ops[i].Dest == "" &&
			ops[i+1].Type == md.OpCodeSpan && ops[i+2].Type == md.OpLinkEnd {

			dest := elvdocTarget(ops[i+1].Text, module)
			ops[i].Dest, ops[i+2].Dest = dest, dest
		}
	}
}

func elvdocTarget(symbol, currentModule string) string {
	i := strings.IndexRune(symbol, ':')
	if i == -1 {
		// An internal link in the builtin module's doc.
		return "#" + symbol
	}

	var module, unqualified string
	if strings.HasPrefix(symbol, "$") {
		module, unqualified = symbol[1:i], "$"+symbol[i+1:]
	} else {
		module, unqualified = symbol[:i], symbol[i+1:]
	}
	switch module {
	case "builtin":
		// A link from a non-builtin module's doc to the builtin module. Use
		// unqualified name (like #put or #$paths, instead of #builtin:put or
		// #$builtin:paths).
		return "builtin.html#" + unqualified
	case currentModule:
		// An internal link in a non-builtin module's doc.
		return "#" + symbol
	default:
		// A link to a non-builtin module.
		return module + ".html#" + symbol
	}
}

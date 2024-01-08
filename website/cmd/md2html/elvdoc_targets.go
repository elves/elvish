package main

import (
	"strings"

	"src.elv.sh/pkg/md"
)

// Adds the implicit destination for [`foo`]().
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

// foo -> builtin.html#foo
// $foo -> builtin.html#$foo
// mod:foo -> mod.html#mod:foo
// $mod:foo -> mod.html#$mod:foo
func elvdocTarget(symbol, currentModule string) string {
	var module string
	i := strings.IndexRune(symbol, ':')
	if i == -1 {
		module = "builtin"
	} else if strings.HasPrefix(symbol, "$") {
		module = symbol[1:i]
	} else {
		module = symbol[:i]
	}

	if module == currentModule {
		return "#" + symbol
	}
	return module + ".html#" + symbol
}

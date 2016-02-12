package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// stylist takes a Node and Editor, and returns a style string. The Node is
// always a leaf in the parsed AST.
// NOTE: Not all stylings are now done with stylists.
type stylist func(parse.Node, *Editor) string

var stylists = []stylist{
	colorFormHead,
}

func colorFormHead(n parse.Node, ed *Editor) string {
	// BUG doesn't work when the form head is compound
	n, head := formHead(n)
	if n == nil {
		return ""
	}
	if goodFormHead(head, ed) {
		return styleForGoodCommand
	}
	return styleForBadCommand
}

func goodFormHead(head string, ed *Editor) bool {
	if isBuiltinSpecial[head] {
		return true
	} else if eval.DontSearch(head) {
		// XXX don't stat twice
		return eval.IsExecutable(head) || isDir(head)
	} else {
		return ed.evaler.Global()[eval.FnPrefix+head] != nil ||
			ed.isExternal[head]
	}
}

var isBuiltinSpecial = map[string]bool{}

func init() {
	for _, name := range eval.BuiltinSpecialNames {
		isBuiltinSpecial[name] = true
	}
}

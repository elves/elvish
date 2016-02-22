package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
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
	} else if util.DontSearch(head) {
		// XXX don't stat twice
		return util.IsExecutable(head) || isDir(head)
	} else {
		return ed.isExternal[head] ||
			ed.evaler.Builtin()[eval.FnPrefix+head] != nil ||
			ed.evaler.Global()[eval.FnPrefix+head] != nil
	}
}

var isBuiltinSpecial = map[string]bool{}

func init() {
	for _, name := range eval.BuiltinSpecialNames {
		isBuiltinSpecial[name] = true
	}
}

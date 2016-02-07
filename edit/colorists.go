package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type colorist func(parse.Node, *Editor) string

var colorists = []colorist{
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
		ed.isExternal.RLock()
		defer ed.isExternal.RUnlock()
		return ed.evaler.Global()[eval.FnPrefix+head] != nil ||
			ed.isExternal.m[head]
	}
}

var isBuiltinSpecial = map[string]bool{}

func init() {
	for _, name := range eval.BuiltinSpecialNames {
		isBuiltinSpecial[name] = true
	}
}

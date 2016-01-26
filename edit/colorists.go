package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type colorist func(parse.Node, *Editor) string

var colorists = []colorist{
	colorFormHead,
	colorVariable,
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
		return eval.IsExecutable(head)
	} else {
		return ed.evaler.Global()["fn-"+head] != nil || ed.isExternal[head]
	}
}

var isBuiltinSpecial = map[string]bool{}

func init() {
	for _, name := range eval.BuiltinSpecialNames {
		isBuiltinSpecial[name] = true
	}
}

func colorVariable(n parse.Node, ed *Editor) string {
	pn, ok := n.(*parse.Primary)
	if !ok {
		return ""
	}
	if pn.Type != parse.Variable || len(pn.Value) == 0 {
		return ""
	}
	if ed.evaler.Global()[pn.Value[1:]] != nil {
		return ""
	}
	return styleForBadVariable
}

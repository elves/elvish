package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type colorist func(parse.Node, *eval.Evaler) string

var colorists = []colorist{
	colorFormHead,
	colorVariable,
}

func colorFormHead(n parse.Node, ev *eval.Evaler) string {
	// BUG doesn't work when the form head is compound
	return ""
}

func colorVariable(n parse.Node, ev *eval.Evaler) string {
	pn, ok := n.(*parse.Primary)
	if !ok {
		return ""
	}
	if pn.Type != parse.Variable || len(pn.Value) == 0 {
		return ""
	}
	has := ev.HasVariable(pn.Value[1:])
	if has {
		return ""
	}
	return styleForBadVariable
}

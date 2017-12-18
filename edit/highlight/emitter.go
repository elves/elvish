package highlight

import (
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type Emitter struct {
	GoodFormHead func(string) bool
	AddStyling   func(begin, end int, style string)
}

func (e *Emitter) EmitAll(n parse.Node) {
	switch n := n.(type) {
	case *parse.Form:
		e.form(n)
	case *parse.Primary:
		e.primary(n)
	case *parse.Sep:
		e.sep(n)
	}
	for _, child := range n.Children() {
		e.EmitAll(child)
	}
}

func (e *Emitter) form(n *parse.Form) {
	for _, an := range n.Assignments {
		if an.Left != nil && an.Left.Head != nil {
			v := an.Left.Head
			e.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
		}
	}
	for _, cn := range n.Vars {
		if len(cn.Indexings) > 0 && cn.Indexings[0].Head != nil {
			v := cn.Indexings[0].Head
			e.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
		}
	}
	if n.Head != nil {
		e.formHead(n.Head)
		// Special forms
		switch n.Head.SourceText() {
		case "if":
			for i := 2; i < len(n.Args); i += 2 {
				a := n.Args[i]
				argText := a.SourceText()
				if argText == "elif" || argText == "else" {
					e.AddStyling(a.Begin(), a.End(), styleForSep[argText])
				}
			}
		case "for":
			if len(n.Args) >= 1 && len(n.Args[0].Indexings) > 0 {
				v := n.Args[0].Indexings[0].Head
				e.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
			if len(n.Args) >= 4 && n.Args[3].SourceText() == "else" {
				a := n.Args[3]
				e.AddStyling(a.Begin(), a.End(), styleForSep["else"])
			}
		case "try":
			i := 1
			highlightKeyword := func(name string) bool {
				if i >= len(n.Args) {
					return false
				}
				a := n.Args[i]
				if a.SourceText() != name {
					return false
				}
				e.AddStyling(a.Begin(), a.End(), styleForSep[name])
				return true
			}
			if highlightKeyword("except") {
				if i+1 < len(n.Args) && len(n.Args[i+1].Indexings) > 0 {
					v := n.Args[i+1].Indexings[0]
					e.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
				}
				i += 3
			}
			if highlightKeyword("else") {
				i += 2
			}
			highlightKeyword("finally")
		}
		// TODO(xiaq): Handle other special forms.
	}
}

func (e *Emitter) formHead(n *parse.Compound) {
	head, err := eval.PurelyEvalCompound(n)
	st := ui.Styles{}
	if err == nil {
		if e.GoodFormHead(head) {
			st = styleForGoodCommand
		} else {
			st = styleForBadCommand
		}
	} else if err != eval.ErrImpure {
		st = styleForBadCommand
	}
	if len(st) > 0 {
		e.AddStyling(n.Begin(), n.End(), st.String())
	}
}

func (e *Emitter) primary(n *parse.Primary) {
	e.AddStyling(n.Begin(), n.End(), styleForPrimary[n.Type].String())
}

func (e *Emitter) sep(n *parse.Sep) {
	septext := n.SourceText()
	switch {
	case strings.TrimSpace(septext) == "":
		// Don't do anything. Whitespaces don't get any styling.
	case strings.HasPrefix(septext, "#"):
		// Comment.
		e.AddStyling(n.Begin(), n.End(), styleForComment.String())
	default:
		e.AddStyling(n.Begin(), n.End(), styleForSep[septext])
	}
}

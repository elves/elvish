package highlight

import (
	"strings"

	"github.com/elves/elvish/edit/nodeutil"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
)

type Highlighter struct {
	GoodFormHead func(string) bool
	AddStyling   func(begin, end int, style string)
}

func (s *Highlighter) Highlight(n parse.Node) {
	switch n := n.(type) {
	case *parse.Form:
		for _, an := range n.Assignments {
			if an.Left != nil && an.Left.Head != nil {
				v := an.Left.Head
				s.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		}
		for _, cn := range n.Vars {
			if len(cn.Indexings) > 0 && cn.Indexings[0].Head != nil {
				v := cn.Indexings[0].Head
				s.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		}
		if n.Head != nil {
			s.formHead(n.Head)
		}
		if n.Head != nil {
			// Special forms
			switch n.Head.SourceText() {
			case "for":
				if len(n.Args) >= 1 && len(n.Args[0].Indexings) > 0 {
					v := n.Args[0].Indexings[0].Head
					s.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
				}
				if len(n.Args) >= 4 && n.Args[3].SourceText() == "else" {
					a := n.Args[3]
					s.AddStyling(a.Begin(), a.End(), styleForSep["else"])
				}
			case "try":
				i := 1
				highlightKeyword := func(name string) bool {
					if i >= len(n.Args) {
						return false
					}
					a := n.Args[i]
					if a.SourceText() == name {
						s.AddStyling(a.Begin(), a.End(), styleForSep[name])
					}
					return true
				}
				if highlightKeyword("except") {
					if i+1 < len(n.Args) && len(n.Args[i+1].Indexings) > 0 {
						v := n.Args[i+1].Indexings[0]
						s.AddStyling(v.Begin(), v.End(), styleForGoodVariable.String())
					}
					i += 3
				}
				if highlightKeyword("else") {
					i += 2
				}
				highlightKeyword("finally")
			}
		}
	case *parse.Primary:
		s.AddStyling(n.Begin(), n.End(), styleForPrimary[n.Type].String())
	case *parse.Sep:
		septext := n.SourceText()
		if strings.HasPrefix(septext, "#") {
			s.AddStyling(n.Begin(), n.End(), styleForComment.String())
		} else {
			s.AddStyling(n.Begin(), n.End(), styleForSep[septext])
		}
	}
	for _, child := range n.Children() {
		s.Highlight(child)
	}
}

func (s *Highlighter) formHead(n *parse.Compound) {
	simple, head, err := nodeutil.SimpleCompound(n, nil)
	st := ui.Styles{}
	if simple {
		if s.GoodFormHead(head) {
			st = styleForGoodCommand
		} else {
			st = styleForBadCommand
		}
	} else if err != nil {
		st = styleForBadCommand
	}
	if len(st) > 0 {
		s.AddStyling(n.Begin(), n.End(), st.String())
	}
}

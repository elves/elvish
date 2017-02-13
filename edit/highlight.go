package edit

import (
	"os"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type Highlighter struct {
	goodFormHead func(string) bool
	addStyling   func(begin, end int, style string)
}

func highlight(n parse.Node, ed *Editor) {
	s := &Highlighter{
		func(s string) bool { return goodFormHead(s, ed) },
		ed.styling.add,
	}
	s.highlight(n)
}

func (s *Highlighter) highlight(n parse.Node) {
	switch n := n.(type) {
	case *parse.Form:
		for _, an := range n.Assignments {
			if an.Dst != nil && an.Dst.Head != nil {
				v := an.Dst.Head
				s.addStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		}
		for _, cn := range n.Vars {
			if len(cn.Indexings) > 0 && cn.Indexings[0].Head != nil {
				v := cn.Indexings[0].Head
				s.addStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		}
		if n.Head != nil {
			s.formHead(n.Head)
		}
	case *parse.Control:
		switch n.Kind {
		case parse.ForControl:
			if n.Iterator != nil {
				v := n.Iterator.Head
				s.addStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		case parse.TryControl:
			if n.ExceptVar != nil {
				v := n.ExceptVar.Head
				s.addStyling(v.Begin(), v.End(), styleForGoodVariable.String())
			}
		}
	case *parse.Primary:
		s.addStyling(n.Begin(), n.End(), styleForPrimary[n.Type].String())
	case *parse.Sep:
		septext := n.SourceText()
		if strings.HasPrefix(septext, "#") {
			s.addStyling(n.Begin(), n.End(), styleForComment.String())
		} else {
			s.addStyling(n.Begin(), n.End(), styleForSep[septext])
		}
	}
	for _, child := range n.Children() {
		s.highlight(child)
	}
}

func (s *Highlighter) formHead(n *parse.Compound) {
	simple, head, err := simpleCompound(n, nil)
	st := styles{}
	if simple {
		if s.goodFormHead(head) {
			st = styleForGoodCommand
		} else {
			st = styleForBadCommand
		}
	} else if err != nil {
		st = styleForBadCommand
	}
	if len(st) > 0 {
		s.addStyling(n.Begin(), n.End(), st.String())
	}
}

func goodFormHead(head string, ed *Editor) bool {
	if isBuiltinSpecial[head] {
		return true
	} else if util.DontSearch(head) {
		// XXX don't stat twice
		return util.IsExecutable(head) || isDir(head)
	} else {
		explode, ns, name := eval.ParseVariable(head)
		if !explode {
			switch ns {
			case "":
				if eval.Builtin()[eval.FnPrefix+name] != nil || ed.evaler.Global[eval.FnPrefix+name] != nil {
					return true
				}
			case "e":
				if ed.isExternal[name] {
					return true
				}
			default:
				if ed.evaler.Modules[ns] != nil && ed.evaler.Modules[ns][eval.FnPrefix+name] != nil {
					return true
				}
			}
		}
		return ed.isExternal[head]
	}
}

var isBuiltinSpecial = map[string]bool{}

func init() {
	for _, name := range eval.BuiltinSpecialNames {
		isBuiltinSpecial[name] = true
	}
}

func isDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && stat.IsDir()
}

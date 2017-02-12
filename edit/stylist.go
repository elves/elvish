package edit

import (
	"os"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type Stylist struct {
	editor *Editor
}

func stylize(n parse.Node, ed *Editor) {
	s := &Stylist{ed}
	s.stylize(n)
}

func (s *Stylist) add(style string, begin, end int) {
	s.editor.styling.add(begin, end, style)
}

func (s *Stylist) stylize(n parse.Node) {
	switch n := n.(type) {
	case *parse.Form:
		for _, an := range n.Assignments {
			if an.Dst != nil && an.Dst.Head != nil {
				v := an.Dst.Head
				s.add(styleForGoodVariable.String(), v.Begin(), v.End())
			}
		}
		for _, cn := range n.Vars {
			if len(cn.Indexings) > 0 && cn.Indexings[0].Head != nil {
				v := cn.Indexings[0].Head
				s.add(styleForGoodVariable.String(), v.Begin(), v.End())
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
				s.add(styleForGoodVariable.String(), v.Begin(), v.End())
			}
		case parse.TryControl:
			if n.ExceptVar != nil {
				v := n.ExceptVar.Head
				s.add(styleForGoodVariable.String(), v.Begin(), v.End())
			}
		}
	case *parse.Primary:
		s.add(styleForPrimary[n.Type].String(), n.Begin(), n.End())
	case *parse.Sep:
		septext := n.SourceText()
		if strings.HasPrefix(septext, "#") {
			s.add(styleForComment.String(), n.Begin(), n.End())
		} else {
			s.add(styleForSep[septext], n.Begin(), n.End())
		}
	}
	for _, child := range n.Children() {
		s.stylize(child)
	}
}

func (s *Stylist) formHead(n *parse.Compound) {
	simple, head, err := simpleCompound(n, nil)
	st := styles{}
	if simple {
		if goodFormHead(head, s.editor) {
			st = styleForGoodCommand
		} else {
			st = styleForBadCommand
		}
	} else if err != nil {
		st = styleForBadCommand
	}
	if len(st) > 0 {
		s.add(st.String(), n.Begin(), n.End())
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

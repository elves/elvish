package edit

import (
	"os"
	"sort"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type Stylist struct {
	tokens   []Token
	editor   *Editor
	commands styleCommands
}

func (s *Stylist) do(n parse.Node) {
	s.style(n)
	s.apply()
}

func (s *Stylist) apply() {
	sort.Sort(s.commands)
	tokens, commands := s.tokens, s.commands

	for len(tokens) > 0 && len(commands) > 0 {
		cmd := commands[0]
		for len(tokens) > 0 && tokens[0].Node.Begin() < cmd.begin {
			tokens = tokens[1:]
		}
		for len(tokens) > 0 && tokens[0].Node.End() <= cmd.end {
			tokens[0].MoreStyle = append(tokens[0].MoreStyle, cmd.style)
			tokens = tokens[1:]
		}
		commands = commands[1:]
	}
}

func (s *Stylist) add(style string, begin, end int) {
	s.commands = append(s.commands, &styleCommand{style, begin, end})
}

func (s *Stylist) style(n parse.Node) {
	if fn, ok := n.(*parse.Form); ok {
		for _, an := range fn.Assignments {
			if an.Dst != nil && an.Dst.Head != nil {
				v := an.Dst.Head
				s.add(styleForType[Variable].String(), v.Begin(), v.End())
			}
		}
		if fn.Head != nil {
			s.formHead(fn.Head)
		}
	}
	if cn, ok := n.(*parse.Control); ok {
		switch cn.Kind {
		case parse.ForControl:
			if cn.Iterator != nil {
				v := cn.Iterator.Head
				s.add(styleForType[Variable].String(), v.Begin(), v.End())
			}
		case parse.TryControl:
			if cn.ExceptVar != nil {
				v := cn.ExceptVar.Head
				s.add(styleForType[Variable].String(), v.Begin(), v.End())
			}
		}
	}
	for _, child := range n.Children() {
		s.style(child)
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

type styleCommand struct {
	style string
	begin int
	end   int
}

type styleCommands []*styleCommand

func (sc styleCommands) Len() int           { return len(sc) }
func (sc styleCommands) Swap(i, j int)      { sc[i], sc[j] = sc[j], sc[i] }
func (sc styleCommands) Less(i, j int) bool { return sc[i].begin < sc[j].begin }

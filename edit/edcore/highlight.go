package edcore

import (
	"os"

	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func doHighlight(n parse.Node, ed *editor) {
	s := &highlight.Emitter{
		func(s string) bool { return goodFormHead(s, ed) },
		ed.styling.Add,
	}
	s.EmitAll(n)
}

func goodFormHead(head string, ed *editor) bool {
	if eval.IsBuiltinSpecial[head] {
		return true
	} else if util.DontSearch(head) {
		// XXX don't stat twice
		return util.IsExecutable(head) || isDir(head)
	} else {
		ev := ed.evaler
		explode, ns, name := eval.ParseVariableRef(head)
		if !explode {
			switch ns {
			case "":
				if ev.Builtin[name+eval.FnSuffix] != nil || ev.Global[name+eval.FnSuffix] != nil {
					return true
				}
			case "e":
				if ed.isExternal[name] {
					return true
				}
			default:
				mod := ev.Global[ns+eval.NsSuffix]
				if mod == nil {
					mod = ev.Builtin[ns+eval.NsSuffix]
				}
				if mod != nil && mod.Get().(eval.Ns)[name+eval.FnSuffix] != nil {
					return true
				}
			}
		}
		return ed.isExternal[head]
	}
}

func isDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && stat.IsDir()
}

package edit

import (
	"os"

	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func doHighlight(n parse.Node, ed *Editor) {
	s := &highlight.Emitter{
		func(s string) bool { return goodFormHead(s, ed) },
		ed.styling.add,
	}
	s.EmitAll(n)
}

func goodFormHead(head string, ed *Editor) bool {
	if eval.IsBuiltinSpecial[head] {
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

func isDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && stat.IsDir()
}

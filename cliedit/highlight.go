package cliedit

import (
	"os"
	"os/exec"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cliedit/highlight"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func initHighlighter(app *cli.App, ev *eval.Evaler) {
	app.Config.Highlighter = highlight.NewHighlighter(highlight.Dep{
		Check:      func(n *parse.Chunk) error { return check(ev, n) },
		HasCommand: func(cmd string) bool { return hasCommand(ev, cmd) },
	})
}

func check(ev *eval.Evaler, n *parse.Chunk) error {
	src := eval.NewInteractiveSource(n.SourceText())
	_, err := ev.Compile(n, src)
	return err
}

func hasCommand(ev *eval.Evaler, cmd string) bool {
	if eval.IsBuiltinSpecial[cmd] {
		return true
	}
	if util.DontSearch(cmd) {
		return isDirOrExecutable(cmd) || hasExternalCommand(cmd)
	}

	sigil, qname := eval.SplitVariableRef(cmd)
	if sigil != "" {
		// The @ sign is only valid when referring to external commands.
		return hasExternalCommand(cmd)
	}

	firstNs, rest := eval.SplitIncompleteQNameFirstNs(qname)
	switch firstNs {
	case "e:":
		return hasExternalCommand(rest)
	case "":
		// Unqualified name; try builtin and global.
		if hasFn(ev.Builtin, rest) || hasFn(ev.Global, rest) {
			return true
		}
	default:
		// Qualified name. Find the top-level module first.
		if hasQualifiedFn(ev, firstNs, rest) {
			return true
		}
	}

	// If all failed, it can still be an external command.
	return hasExternalCommand(cmd)
}

func hasQualifiedFn(ev *eval.Evaler, firstNs string, rest string) bool {
	if rest == "" {
		return false
	}
	modVar := ev.Global[firstNs]
	if modVar == nil {
		modVar = ev.Builtin[firstNs]
		if modVar == nil {
			return false
		}
	}
	mod, ok := modVar.Get().(eval.Ns)
	if !ok {
		return false
	}
	segs := eval.SplitQNameNsSegs(rest)
	for _, seg := range segs[:len(segs)-1] {
		modVar = mod[seg]
		if modVar == nil {
			return false
		}
		mod, ok = modVar.Get().(eval.Ns)
		if !ok {
			return false
		}
	}
	return hasFn(mod, segs[len(segs)-1])
}

func hasFn(ns eval.Ns, name string) bool {
	fnVar, ok := ns[name+eval.FnSuffix]
	if !ok {
		return false
	}
	_, ok = fnVar.Get().(eval.Callable)
	return ok
}

func isDirOrExecutable(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && (stat.IsDir() || stat.Mode()&0111 != 0)
}

func hasExternalCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

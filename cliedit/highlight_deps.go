package cliedit

import (
	"os"
	"os/exec"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func makeCheck(ev *eval.Evaler) func(n *parse.Chunk) error {
	return func(n *parse.Chunk) error { return check(ev, n) }
}

func check(ev *eval.Evaler, n *parse.Chunk) error {
	src := eval.NewInteractiveSource(n.SourceText())
	_, err := ev.Compile(n, src)
	return err
}

func makeHasCommand(ev *eval.Evaler) func(cmd string) bool {
	return func(cmd string) bool { return hasCommand(ev, cmd) }
}

func hasCommand(ev *eval.Evaler, cmd string) bool {
	if eval.IsBuiltinSpecial[cmd] {
		return true
	}
	if util.DontSearch(cmd) {
		return isDirOrExecutable(cmd) || hasExternalCommand(cmd)
	}

	explode, ns, name := eval.ParseVariableRef(cmd)
	if explode {
		// The @ sign is only valid when referring to external commands.
		return hasExternalCommand(cmd)
	}

	switch ns {
	case "e":
		return hasExternalCommand(name)
	case "":
		// Unqualified name; try builtin and global.
		if hasFn(ev.Builtin, name) || hasFn(ev.Global, name) {
			return true
		}
	default:
		// Qualified name. Find the top-level module first.
		if hasQualifiedFn(ev, strings.Split(ns, ":"), name) {
			return true
		}
	}

	// If all failed, it can still be an external command.
	return hasExternalCommand(cmd)
}

func hasQualifiedFn(ev *eval.Evaler, nsParts []string, name string) bool {
	modVar := ev.Global[nsParts[0]+eval.NsSuffix]
	if modVar == nil {
		modVar = ev.Builtin[nsParts[0]+eval.NsSuffix]
		if modVar == nil {
			return false
		}
	}
	mod, ok := modVar.Get().(eval.Ns)
	if !ok {
		return false
	}
	for _, nsPart := range nsParts[1:] {
		modVar = mod[nsPart+eval.NsSuffix]
		if modVar == nil {
			return false
		}
		mod, ok = modVar.Get().(eval.Ns)
		if !ok {
			return false
		}
	}
	return hasFn(mod, name)
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

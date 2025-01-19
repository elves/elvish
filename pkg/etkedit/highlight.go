package edit

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// Declared as global variables to be swapped out in tests.
var (
	hasCommand         = hasCommandImpl
	hasCommandMaxBlock = func() time.Duration { return 10 * time.Millisecond }
)

func getHighlighter(ev *eval.Evaler) *highlight.Highlighter {
	return highlight.NewHighlighter(highlight.Config{
		Check: func(t parse.Tree) (string, []*eval.CompilationError) {
			autofixes, err := ev.CheckTree(t, nil)
			autofix := strings.Join(autofixes, "; ")
			return autofix, eval.UnpackCompilationErrors(err)
		},
		HasCommand:         func(cmd string) bool { return hasCommand(ev, cmd) },
		HasCommandMaxBlock: func() time.Duration { return hasCommandMaxBlock() },
		AutofixTip: func(autofix string) ui.Text {
			return ui.T("")
			/*
				return bindingTips(ed.ns, "insert:binding",
					bindingTip("autofix: "+autofix, "apply-autofix"),
				    bindingTip("autofix first", "smart-enter",
				    "completion:smart-start"))
			*/
		},
	})
}

// TODO: Below is duplicate with pkg/edit/highlight.go

func hasCommandImpl(ev *eval.Evaler, cmd string) bool {
	if eval.IsBuiltinSpecial[cmd] {
		return true
	}
	if fsutil.DontSearch(cmd) {
		return isDirOrExecutable(cmd) || hasExternalCommand(cmd)
	}

	sigil, qname := eval.SplitSigil(cmd)
	if sigil != "" {
		// The @ sign is only valid when referring to external commands.
		return hasExternalCommand(cmd)
	}

	first, rest := eval.SplitQName(qname)
	switch {
	case rest == "":
		// Unqualified name; try builtin and global.
		if hasFn(ev.Builtin(), first) || hasFn(ev.Global(), first) {
			return true
		}
	case first == "e:":
		return hasExternalCommand(rest)
	default:
		// Qualified name. Find the top-level module first.
		if hasQualifiedFn(ev, first, rest) {
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
	modVal, ok := ev.Global().Index(firstNs)
	if !ok {
		modVal, ok = ev.Builtin().Index(firstNs)
		if !ok {
			return false
		}
	}
	mod, ok := modVal.(*eval.Ns)
	if !ok {
		return false
	}
	segs := eval.SplitQNameSegs(rest)
	for _, seg := range segs[:len(segs)-1] {
		modVal, ok = mod.Index(seg)
		if !ok {
			return false
		}
		mod, ok = modVal.(*eval.Ns)
		if !ok {
			return false
		}
	}
	return hasFn(mod, segs[len(segs)-1])
}

func hasFn(ns *eval.Ns, name string) bool {
	fnVar, ok := ns.Index(name + eval.FnSuffix)
	if !ok {
		return false
	}
	_, ok = fnVar.(eval.Callable)
	return ok
}

func isDirOrExecutable(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && (stat.IsDir() || fsutil.IsExecutable(stat))
}

func hasExternalCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

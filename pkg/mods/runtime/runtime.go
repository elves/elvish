// Package runtime implements the runtime module.
package runtime

import (
	"os"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

var osExecutable = os.Executable

// Ns returns the namespace for the runtime: module.
//
// All the public properties of the Evaler should be set before this function is
// called.
func Ns(ev *eval.Evaler) *eval.Ns {
	elvishPath, err := osExecutable()
	if err != nil {
		elvishPath = ""
	}

	return eval.BuildNsNamed("runtime").
		AddVars(map[string]vars.Var{
			"elvish-path":       vars.NewReadOnly(nonEmptyOrNil(elvishPath)),
			"lib-dirs":          vars.NewReadOnly(vals.MakeListSlice(ev.LibDirs)),
			"rc-path":           vars.NewReadOnly(nonEmptyOrNil(ev.RcPath)),
			"effective-rc-path": vars.NewReadOnly(nonEmptyOrNil(ev.EffectiveRcPath)),
		}).Ns()
}

func nonEmptyOrNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

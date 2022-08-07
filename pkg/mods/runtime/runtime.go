// Package runtime implements the runtime module.
package runtime

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

//elvdoc:var lib-dirs
//
// A list containing
// [module search directories](command.html#module-search-directories).

//elvdoc:var rc-path
//
// Path to the [RC file](command.html#rc-file), ignoring any possible overrides
// by command-line flags and available in non-interactive mode.
//
// If there was an error in determining the path of the RC file, this variable
// is `$nil`.
//
// @cf $runtime:effective-rc-path

//elvdoc:var effective-rc-path
//
// Path to the [RC path](command.html#rc-file) that is actually used for this
// Elvish session:
//
// - If Elvish is running non-interactively or invoked with the `-norc` flag,
//   this variable is `$nil`.
//
// - If Elvish is invoked with the `-rc` flag, this variable contains the
//   absolute path of the argument.
//
// - Otherwise (when Elvish is running interactively and invoked without
//   `-norc` or `-rc`), this variable has the same value as `$rc-path`.
//
// @cf $runtime:rc-path

// Ns returns the namespace for the runtime: module.
//
// All the public properties of the Evaler should be set before this function is
// called.
func Ns(ev *eval.Evaler) *eval.Ns {
	return eval.BuildNsNamed("runtime").
		AddVars(map[string]vars.Var{
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

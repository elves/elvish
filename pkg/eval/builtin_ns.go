package eval

import (
	"strconv"
	"syscall"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/eval/vars"
)

//elvdoc:var _
//
// A blackhole variable.
//
// Values assigned to it will be discarded. Referencing it always results in $nil.

//elvdoc:var args
//
// A list containing command-line arguments. Analogous to `argv` in some other
// languages. Examples:
//
// ```elvish-transcript
// ~> echo 'put $args' > args.elv
// ~> elvish args.elv foo -bar
// ▶ [foo -bar]
// ~> elvish -c 'put $args' foo -bar
// ▶ [foo -bar]
// ```
//
// As demonstrated above, this variable does not contain the name of the script
// used to invoke it. For that information, use the `src` command.
//
// @cf src

//elvdoc:var false
//
// The boolean false value.

//elvdoc:var ok
//
// The special value used by `?()` to signal absence of exceptions.

//elvdoc:var nil
//
// A special value useful for representing the lack of values.

//elvdoc:var paths
//
// A list of search paths, kept in sync with `$E:PATH`. It is easier to use than
// `$E:PATH`.

//elvdoc:var pid
//
// The process ID of the current Elvish process.

//elvdoc:var pwd
//
// The present working directory. Setting this variable has the same effect as
// `cd`. This variable is most useful in a temporary assignment.
//
// Example:
//
// ```elvish
// ## Updates all git repositories
// for x [*/] {
//   pwd=$x {
//     if ?(test -d .git) {
//       git pull
//     }
//   }
// }
// ```
//
// Etymology: the `pwd` command.
//
// @cf cd

//elvdoc:var true
//
// The boolean true value.

//elvdoc:var buildinfo
//
// A [psuedo-map](./language.html#pseudo-map) that exposes information about the Elvish binary.
// Running `put $buildinfo | to-json` will produce the same output as `elvish -buildinfo -json`.
//
// @cf version

//elvdoc:var version
//
// The full version of the Elvish binary as a string. This is the same information reported by
// `elvish -version` and the value of `$buildinfo[version]`.
//
// **Note:** In general it is better to perform functionality tests rather than testing `$version`.
// For example, do something like
//
// ```
// has-key $builtin: new-var
// ````
//
// to test if variable `new-var` is available rather than comparing against `$version` to see if the
// elvish version is equal to or newer than the version that introduced `new-var`.
//
// @cf buildinfo

var builtinNs = BuildNsNamed("").AddVars(map[string]vars.Var{
	"_":         vars.NewBlackhole(),
	"pid":       vars.NewReadOnly(strconv.Itoa(syscall.Getpid())),
	"ok":        vars.NewReadOnly(OK),
	"nil":       vars.NewReadOnly(nil),
	"true":      vars.NewReadOnly(true),
	"false":     vars.NewReadOnly(false),
	"buildinfo": vars.NewReadOnly(buildinfo.Value),
	"version":   vars.NewReadOnly(buildinfo.Value.Version),
	"paths":     vars.NewEnvListVar("PATH"),
})

func addBuiltinFns(fns map[string]interface{}) {
	builtinNs.AddGoFns(fns)
}

package eval

import (
	"strconv"
	"syscall"

	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/xiaq/persistent/vector"
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

var builtinNs = Ns{
	"_":     vars.NewBlackhole(),
	"args":  vars.NewReadOnly("args", vector.Empty),
	"pid":   vars.NewReadOnly("pid", strconv.Itoa(syscall.Getpid())),
	"ok":    vars.NewReadOnly("ok", OK),
	"nil":   vars.NewReadOnly("nil", nil),
	"true":  vars.NewReadOnly("true", true),
	"false": vars.NewReadOnly("false", false),
	"paths": NewEnvListVar("PATH"),
}

func addBuiltinFns(fns map[string]interface{}) {
	builtinNs.AddGoFns("", fns)
}

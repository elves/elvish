// Package exc exposes an Elvish module containing functionalities for working
// with exceptions.
package exc

import "github.com/elves/elvish/pkg/eval"

// Ns contains the exc: namespace.
var Ns = eval.Ns{}.
	AddGoFns("exc:", map[string]interface{}{
		"show": show,

		"is-external-cmd-exc": isExternalCmdExc,
		"is-nonzero-exit":     isNonzeroExit,
		"is-killed":           isKilled,

		"is-fail-exc": isFailExc,
	})

//elvdoc:fn show
//
// ```elvish
// exc:show $e
// ```
//
// Prints the exception to the output, showing its cause and stacktrace using VT
// sequences.
//
// Example:
//
// ```elvish-transcript
// ~> e = ?(fail lorem-ipsum)
// ~> exc:show $e
// Exception: lorem-ipsum
// [tty 3], line 1: e = ?(fail lorem-ipsum)
// ```

func show(fm *eval.Frame, e *eval.Exception) {
	fm.OutputFile().WriteString(e.Show(""))
	fm.OutputFile().WriteString("\n")
}

//elvdoc:fn is-external-cmd-exc
//
// ```elvish
// exc:is-external-cmd-exc $e
// ```
//
// Outputs whether an exception was caused by any error with an external
// command. If this command outputs `$true`, exact one of `exc:is-nonzero-exit
// $e` and `exc:is-killed $e` will output `$true`.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-external-cmd-exc ?(fail bad)
// ▶ $false
// ~> exc:is-external-cmd-exc ?(false)
// ▶ $true
// ~> exc:is-external-cmd-exc ?(elvish -c 'echo $pid; exec cat')
// # outputs pid
// # run "kill <pid> in another terminal"
// ▶ $true
// ```
//
// @cf is-nonzero-exit is-killed

func isExternalCmdExc(e *eval.Exception) bool {
	_, ok := e.Cause.(eval.ExternalCmdExit)
	return ok
}

//elvdoc:fn is-nonzero-exit
//
// ```elvish
// exc:is-nonzero-exit $e
// ```
//
// Outputs whether an exception was caused by an external command exiting with a
// nonzero code.
//
// **NOTE**: An external command is only considered to have exited if it
// terminated on its own. If an exception was caused by an external command
// being killed by a signal, this predicate will output `$false`.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-nonzero-exit ?(fail bad)
// ▶ $false
// ~> exc:is-nonzero-exit ?(false)
// ▶ $true
// ```
//
// @cf is-external-cmd-exc is-killed

func isNonzeroExit(e *eval.Exception) bool {
	err, ok := e.Cause.(eval.ExternalCmdExit)
	return ok && err.Exited()
}

//elvdoc:fn is-killed
//
// ```elvish
// exc:is-killed $e
// ```
//
// Outputs whether an exception was caused by an external command being killed.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-killed ?(fail bad)
// ▶ $false
// ~> exc:is-killed ?(elvish -c 'echo $pid; exec cat')
// # outputs pid
// # run "kill <pid> in another terminal"
// ▶ $true
// ```
//
// @cf is-external-cmd-exc is-nonzero-exit

func isKilled(e *eval.Exception) bool {
	err, ok := e.Cause.(eval.ExternalCmdExit)
	return ok && err.Signaled()
}

//elvdoc:fn is-fail-exc
//
// ```elvish
// exc:is-fail-exc $e
// ```
//
// Outputs whether an exception was thrown by the `fail` command.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-fail-exc ?(fail bad)
// ▶ $true
// ~> exc:is-fail-exc ?(false)
// ▶ $false
// ```
//
// @cf builtin:fail

func isFailExc(e *eval.Exception) bool {
	_, ok := e.Cause.(eval.FailError)
	return ok
}

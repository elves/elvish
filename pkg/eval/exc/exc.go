// Package exc exposes an Elvish module containing functionalities for working
// with exceptions.
package exc

import "github.com/elves/elvish/pkg/eval"

// Ns contains the exc: namespace.
var Ns = eval.Ns{}.
	AddGoFns("exc:", map[string]interface{}{
		"show": show,

		"is-external-cmd-error": isExternalCmdError,
		"is-nonzero-exit":       isNonzeroExit,
		"is-killed":             isKilled,

		"is-fail-error": isFailError,
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

//elvdoc:fn is-external-cmd-error
//
// ```elvish
// exc:is-external-cmd-error $e
// ```
//
// Outputs whether an exception was caused by any error with an external
// command. If this command outputs `$true`, exact one of `exc:is-nonzero-exit
// $e` and `exc:is-killed $e` will output `$true`.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-external-cmd-error ?(fail bad)
// ▶ $false
// ~> exc:is-external-cmd-error ?(false)
// ▶ $true
// ~> exc:is-external-cmd-error ?(elvish -c 'echo $pid; exec cat')
// # outputs pid
// # run "kill <pid> in another terminal"
// ▶ $true
// ```
//
// @cf is-nonzero-exit is-killed

func isExternalCmdError(e *eval.Exception) bool {
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
// @cf is-external-cmd-error is-killed

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
// @cf is-external-cmd-error is-nonzero-exit

func isKilled(e *eval.Exception) bool {
	err, ok := e.Cause.(eval.ExternalCmdExit)
	return ok && err.Signaled()
}

//elvdoc:fn is-fail-error
//
// ```elvish
// exc:is-fail-error $e
// ```
//
// Outputs whether an exception was thrown by the `fail` command.
//
// Examples:
//
// ```elvish-transcript
// ```
//
// @cf builtin:fail

func isFailError(e *eval.Exception) bool {
	_, ok := e.Cause.(eval.FailError)
	return ok
}

// Package exc exposes an Elvish module containing functionalities for working
// with exceptions.
package exc

import "github.com/elves/elvish/pkg/eval"

// Ns contains the exc: namespace.
var Ns = eval.Ns{}.
	AddGoFns("exc:", map[string]interface{}{
		"show": show,

		"is-external-cmd-err": isExternalCmdErr,
		"is-nonzero-exit":     isNonzeroExit,
		"is-killed":           isKilled,

		"is-fail-err": isFailErr,

		"is-pipeline-err": isPipelineErr,
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

//elvdoc:fn is-external-cmd-err
//
// ```elvish
// exc:is-external-cmd-err $r
// ```
//
// Outputs whether an exception reason is an error with an external command.
// If this command outputs `$true`, exact one of `exc:is-nonzero-exit $r` and
// `exc:is-killed $r` will output `$true`.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-external-cmd-exc ?(fail bad)[reason]
// ▶ $false
// ~> exc:is-external-cmd-exc ?(false)[reason]
// ▶ $true
// ~> exc:is-external-cmd-exc ?(elvish -c 'echo $pid; exec cat')[reason]
// # outputs pid
// # run "kill <pid> in another terminal"
// ▶ $true
// ```
//
// @cf is-nonzero-exit is-killed

func isExternalCmdErr(e error) bool {
	_, ok := e.(eval.ExternalCmdExit)
	return ok
}

//elvdoc:fn is-nonzero-exit
//
// ```elvish
// exc:is-nonzero-exit $e
// ```
//
// Outputs whether an exception reason is an external command exiting with a
// nonzero code.
//
// **NOTE**: An external command is only considered to have exited if it
// terminated on its own. If an exception was caused by an external command
// being killed by a signal, this predicate will output `$false` on its reason.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-nonzero-exit ?(fail bad)[reason]
// ▶ $false
// ~> exc:is-nonzero-exit ?(false)[reason]
// ▶ $true
// ```
//
// @cf is-external-cmd-exc is-killed

func isNonzeroExit(e error) bool {
	err, ok := e.(eval.ExternalCmdExit)
	return ok && err.Exited()
}

//elvdoc:fn is-killed
//
// ```elvish
// exc:is-killed $e
// ```
//
// Outputs whether an exception reason is an external command being killed.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-killed ?(fail bad)[reason]
// ▶ $false
// ~> exc:is-killed ?(elvish -c 'echo $pid; exec cat')[reason]
// # outputs pid
// # run "kill <pid> in another terminal"
// ▶ $true
// ```
//
// @cf is-external-cmd-exc is-nonzero-exit

func isKilled(e error) bool {
	err, ok := e.(eval.ExternalCmdExit)
	return ok && err.Signaled()
}

//elvdoc:fn is-fail-err
//
// ```elvish
// exc:is-fail-err $e
// ```
//
// Outputs whether an exception reason originates from the `fail` command.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-fail-exc ?(fail bad)[reason]
// ▶ $true
// ~> exc:is-fail-exc ?(false)[reason]
// ▶ $false
// ```
//
// @cf builtin:fail

func isFailErr(e error) bool {
	_, ok := e.(eval.FailError)
	return ok
}

//elvdoc:fn is-pipeline-err
//
// ```elvish
// exc:is-pipeline-err $r
// ```
//
// Outputs whether an exception reason is a result of multiple commands in a
// pipeline throwing out exceptions.
//
// Examples:
//
// ```elvish-transcript
// ~> exc:is-pipeline-err ?(fail bad)[reason]
// ▶ $false
// ~> exc:is-pipeline-err ?(fail 1 | fail 2)[reason]
// ▶ $true
// ```

func isPipelineErr(e error) bool {
	_, ok := e.(eval.PipelineError)
	return ok
}

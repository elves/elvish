package edit

// This file encapsulates functionality related to a complete REPL cycle. Such as capturing
// information about the most recently executed interactive command.

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

//elvdoc:var after-command
//
// A list of functions to call after each interactive command completes. There is one pre-defined
// function used to populate the [`$edit:command-duration`](./edit.html#editcommand-duration)
// variable. Each function is called with a single [map](https://elv.sh/ref/language.html#map)
// argument containing the following keys:
//
// * `code`: A string containing the code that was executed. If `is-file` is true this is the
// content of the file that was sourced.
//
// * `duration`: A [floating-point number](https://elv.sh/ref/language.html#number) representing the
// command execution duration in seconds.
//
// * `error`: An [exception](../ref/language.html#exception) object if the command terminated with
// an exception, else [`$nil`](../ref/language.html#nil).
//
// * `name`: A string describing where the code originated; e.g., `[tty 1]` for the first
// interactive REPL cycle. If `is-file` is true this is the path of the file that was sourced.
//
// * `is-file`: False if the code was entered interactively. At the moment this is only true when
// this hook is run after sourcing the RC file.
//
// @cf edit:command-duration

//elvdoc:var command-duration
//
// Duration, in seconds, of the most recent interactive command. This can be useful in your prompt
// to provide feedback on how long a command took to run. The initial value of this variable is the
// time to evaluate your *~/.elvish/rc.elv* script before printing the first prompt.
//
// @cf edit:after-command

var commandDuration float64

func initRepl(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	// TODO: Change this to a read-only var, possibly by introducing a vars.FromPtrReadonly
	// function, to guard against attempts to modify the value from Elvish code.
	nb.Add("command-duration", vars.FromPtr(&commandDuration))

	hook := newListVar(vals.EmptyList)
	nb["after-command"] = hook
	ed.AfterCommand = append(ed.AfterCommand,
		func(src parse.Source, duration float64, err error) {
			m := vals.MakeMap("name", src.Name, "code", src.Code, "is-file", src.IsFile,
				"duration", duration, "error", err)
			callHooks(ev, "$<edit>:after-command", hook.Get().(vals.List), m)
		})
}

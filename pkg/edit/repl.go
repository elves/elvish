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
// function used to populate the [`$edit:command-duration`](./edit.html#edit:command-duration)
// variable. Each function is called with a single [map](https://elv.sh/ref/language.html#map)
// argument containing the following keys:
//
// * `src`: Information about the source that was executed, same as what
//   [`src`](builtin.html#src) would output inside the code.
//
// * `duration`: A [floating-point number](https://elv.sh/ref/language.html#number) representing the
// command execution duration in seconds.
//
// * `error`: An [exception](../ref/language.html#exception) object if the command terminated with
// an exception, else [`$nil`](../ref/language.html#nil).
//
// @cf edit:command-duration

//elvdoc:var command-duration
//
// Duration, in seconds, of the most recent interactive command. This can be useful in your prompt
// to provide feedback on how long a command took to run. The initial value of this variable is the
// time to evaluate your [`rc.elv`](command.html#rc-file) before printing the first prompt.
//
// @cf edit:after-command

func initRepl(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	var commandDuration float64
	// TODO: Ensure that this variable can only be written from the Elvish code
	// in elv_init.go.
	nb.AddVar("command-duration", vars.FromPtr(&commandDuration))

	afterCommandHook := newListVar(vals.EmptyList)
	nb.AddVar("after-command", afterCommandHook)
	ed.AfterCommand = append(ed.AfterCommand,
		func(src parse.Source, duration float64, err error) {
			m := vals.MakeMap("src", src, "duration", duration, "error", err)
			callHooks(ev, "$<edit>:after-command", afterCommandHook.Get().(vals.List), m)
		})
}

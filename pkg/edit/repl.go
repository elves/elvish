package edit

// This file encapsulates functionality related to a complete REPL cycle. Such as capturing
// information about the most recently executed interactive command.

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

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
			eval.CallHook(ev, nil, "$<edit>:after-command", afterCommandHook.Get().(vals.List), m)
		})
}
